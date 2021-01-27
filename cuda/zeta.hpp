#pragma once

#include <cuda_runtime.h>
#include <math_constants.h>
#include <iostream>
#include <vector>
#include <cstring>

#include "bitmap.hpp"

#define I complex(0, 1)

#define MIN_N 100
#define MAX_N 1000000
#define TWO_PI (CUDART_PI * 2)
#define SQRT_TWO_PI CUDART_SQRT_2PI
#define CABS_Z_MAX 10000.0
#define MAX_ITS 2000
#define MAX_GAMMA 450
#define EPSILON 10e-15

typedef double real;
typedef unsigned char byte;

struct complex {
    real re;
    real im;

    __host__ __device__
    complex () : re(0), im(0) {}
    __host__ __device__
    complex (const real re) : re(re), im(0) {}
    __host__ __device__
    complex (const real re, const real im) : re(re), im(im) {}
    __host__ __device__
    complex (const complex& c) : re(c.re), im(c.im) {}


    __host__ __device__
    complex operator-() const {
        return {-re, -im};
    }

    __host__ __device__
    complex operator+(const complex& that) const {
        return {re + that.re, im + that.im};
    }
    __host__ __device__ complex operator+(real that) const { return *this + complex(that); }

    __host__ __device__
    complex operator-(const complex& that) const {
        return {re - that.re, im - that.im};
    }
    __host__ __device__ complex operator-(real that) const { return *this - complex(that); }

    __host__ __device__
    complex operator*(const complex& that) const {
        return {
            (re * that.re) - (im * that.im),
            (re * that.im) + (im * that.re)
        };
    }
    __host__ __device__ complex operator*(real that) const { return *this * complex(that); }

    __host__ __device__
    complex operator/(const complex& that) const {
        return {
            ((re * that.re) + (im * that.im)) / ((that.re * that.re) + (that.im * that.im)),
            ((im * that.re) - (re * that.im)) / ((that.re * that.re) + (that.im * that.im))
        };
    }
    __host__ __device__ complex operator/(real that) const { return *this / complex(that); }

    __device__
    real arg() const;
    __device__
    real mod() const;
    __device__
    static complex ccos(complex const& c);
    __device__
    static complex cexp(complex const& c);
    __device__
    static complex clog(complex const& c);
    __device__
    static complex cpow(real x, complex const& exp);
    __device__
    static complex cpow(complex const& x, complex const& exp);

};

struct Settings {
    complex min;
    complex max;
    real res;

    Settings(const real rmin, const real rmax, const real imin, const real imax, const int res)
        : min(rmin, imin), max(rmax, imax), res(res) {}
    Settings(const complex& min, const complex& max, const real res)
        : min(min), max(max), res(res) {}
    Settings(const Settings& settings)
        : min(settings.min), max(settings.max), res(settings.res) {}
    Settings()
        : min(), max(), res(0) {}
};

std::ostream& operator<<(std::ostream &out, const complex &z);
std::ostream& operator<<(std::ostream &out, const Settings &settings);

__device__
complex zeta(complex s);

__device__
complex ems(complex s);

__device__
complex pochhammer(complex s, int n);

__device__
complex complex_gamma(complex s);

__device__
int iterate(complex s, real epsilon);

__global__
void zeta_kernel(Settings settings, int width, int height, byte *pixels);

__global__
void calc_iterations(complex min, complex max, uint size, uint *data);

__host__
void write_image(Settings settings);

__host__
void write_image_(int width, int height, uint *data);

__host__
void iterate_single(complex z);


struct ZetaLUT {

    Settings settings;
    int width;
    int height;
    ZetaLUT *next;

    private:
        byte *pixels;

    public:
        __host__
        ZetaLUT(const Settings& settings, const int width, const int height, byte *pixels)
            : settings(settings), width(width), height(height), pixels(pixels), next(NULL) {}
        __host__
        ZetaLUT(const Settings& settings, const bitmap_image& bitmap) 
            : settings(settings), width(bitmap.width()), height(bitmap.height()), next(NULL) {
            cudaMallocManaged(&pixels, width * height * 3);
            for (int i = 0; i < width * height; i ++) {
                int x = i % width;
                int y = height - (i / width) - 1;
                rgb_t pixel = bitmap.get_pixel(x, y);
                pixels[i * 3 + 0] = pixel.red;
                pixels[i * 3 + 1] = pixel.green;
                pixels[i * 3 + 2] = pixel.blue;
            }
        }
        __host__
        void dispose() {
            cudaFree(pixels);
        }

        __host__ __device__
        int get_pixel(int x, int y, rgb_t *pixel) const {
            if (x >= 0 && x < width && y >= 0 && y < height) {
                int i = x + y * width;
                *pixel = {
                    pixels[i * 3 + 0],
                    pixels[i * 3 + 1],
                    pixels[i * 3 + 2]
                };
                return 1;
            }
            else {
                return 0;
            }
        }

        __host__ __device__
        int lookup(const complex z, rgb_t *color) const {
            real u = (z.re - settings.min.re) / (settings.max.re - settings.min.re);
            real v = (z.im - settings.min.im) / (settings.max.im - settings.min.im);
            int x = u * width;
            int y = v * height;
            return get_pixel(x, y, color);
        }
};

class ZetaLUTCollection {

    private:
        ZetaLUT *head;
        int count;

    public:
        __host__
        ZetaLUTCollection() : count(0), head(NULL) {}
        ~ZetaLUTCollection() {
            ZetaLUT *curr = head;
            ZetaLUT *next = NULL;
            while (curr != NULL) {
                next = curr->next;
                curr->dispose();
                cudaFree(curr);
                curr = next;
            }
        }

        __host__
        void add_lut_from_bitmap(const Settings& settings, const char *fname) {
            bitmap_image image(fname);
            if (!image) {
                std::cerr << "Could not open image: " << fname << std::endl;
                std::exit(1);
            }
            add_lut(ZetaLUT(settings, image));
            std::cout << "Loaded LUT " << fname << std::endl;
        }

        __device__
        int lookup(const complex z, rgb_t *color) const {
            ZetaLUT *curr = head;
            while (curr != NULL) {
                if (curr->lookup(z, color)) {
                    return 1;
                }
                curr = curr->next;
            }
            return 0;
        }

    private:
        __host__
        void add_lut(const ZetaLUT &lut) {
            ZetaLUT *newlut;
            cudaMallocManaged(&newlut, sizeof(ZetaLUT));
            *newlut = lut;
            newlut->next = NULL;
            count ++;
            if (head == NULL) {
                head = newlut;
            }
            else {
                if (head->settings.res < newlut->settings.res) {
                    newlut->next = head;
                    head = newlut;
                }
                else {
                    ZetaLUT *curr = head;
                    while (curr != NULL) {
                        if (curr->settings.res < newlut->settings.res || curr->next == NULL) {
                            newlut->next = curr->next;
                            curr->next = newlut;
                            break;
                        }
                        curr = curr->next;
                    }
                }
            }
        }


};

__global__
void zeta_lut_kernel(Settings settings, int width, int height, byte *pixels, ZetaLUTCollection *luts);

__host__
void write_image_lut(Settings settings, const ZetaLUTCollection& luts);