#include <iostream>
#include <chrono>
#include "bitmap.hpp"
#include "zeta.hpp"

using sysclock = std::chrono::system_clock;
using sec = std::chrono::duration<double>;

extern "C" {
    //
    // size -   We always generate a square patch of data so width and height are
    //          the same.  Size is the size in one direction. The total data 
    //          generated is size * size so len(data) == size * size
    //
    // data -   Pre-allocated buffer of length size*size
    void generate(double minR, double maxR, double minI, double maxI, uint size, uint* data) {
        int procCount;
        int devId;
        cudaGetDevice(&devId);
        cudaDeviceGetAttribute(&procCount, cudaDevAttrMultiProcessorCount, devId);

        int threadsPerBlock;
        cudaDeviceGetAttribute(&threadsPerBlock, cudaDevAttrMaxThreadsPerBlock, devId);

        int grids;
        grids = (size*size + threadsPerBlock - 1) / threadsPerBlock;

        uint* gpu_buffer;
        uint memsize = size * size * sizeof(uint);
        cudaMalloc((void**)&gpu_buffer, memsize);

        complex min = complex(minR, minI);
        complex max = complex(maxR, maxI);

        calc_iterations<<<grids, threadsPerBlock>>>(min, max, size, gpu_buffer);

        cudaDeviceSynchronize();
        cudaMemcpy(data, gpu_buffer, memsize, cudaMemcpyDeviceToHost);
        cudaFree(gpu_buffer);

        write_image_(size, size, data);
    }
}

template<class T>
void prompt(const char *prompt, T *result) {
    std::cout << prompt << ":" << std::endl;
    std::cout << "> ";
    std::cin >> *result;
}

template<class T>
int menu(const char *title, T *options, int count) {
    for(;;) {
        std::cout << title << std::endl;
        for (int i = 0; i < count; i ++) {
            std::cout << i << ": " << options[i] << std::endl;
        }
        std::cout << "> ";
        int selected;
        std::cin  >> selected;
        if (selected < 0 || selected >= count) {
            selected = -1;
            std::cout << "No such option" << std::endl << std::endl;
        }
        else {
            return selected;
        }
    }
}

int main(void) {
    int procCount;
    int devId;
    cudaGetDevice(&devId);
    cudaDeviceGetAttribute(&procCount, cudaDevAttrMultiProcessorCount, devId);
    int threadsPerBlock;
    cudaDeviceGetAttribute(&threadsPerBlock, cudaDevAttrMaxThreadsPerBlock, devId);

    std::cout << "proc count: " << procCount << std::endl;
    std::cout << "threads per block: " << threadsPerBlock << std::endl;

    Settings settings;

    uint size;
    prompt<double>("Real Min:", &settings.min.re);
    prompt<double>("Real Max:", &settings.max.re);
    prompt<double>("Imaginary Min:", &settings.min.im);
    prompt<double>("Imaginary Max:", &settings.max.im);
    prompt<uint>("Size (pixels wide):", &size);

    const auto before = sysclock::now();
    // write_image(settings);
    uint* data;
    data = (uint*)malloc(size*size * sizeof(uint));
    generate(
        settings.min.re,
        settings.max.re, 
        settings.min.im,
        settings.max.im, 
        size, data);

    free(data);
    const sec duration = sysclock::now() - before;
    std::cout << "It took " << duration.count() << "s" << std::endl;
    return 0;
}

