#include <math.h>
#include <iostream>
#include <math_constants.h>
#include <cuda_runtime.h>
#include "zeta.hpp"
#include "bitmap.hpp"

__device__
int cc[300] = { 0, 0, 0, 0, 0, 0, 0, 0, 255, 0, 60, 255, 0, 100, 255, 0, 125, 255, 0, 140, 255, 0, 155, 255, 0, 170, 255, 0, 180, 255, 0, 190, 255, 0, 200, 255, 0, 210, 255, 0, 220, 255, 0, 225, 255, 0, 230, 255, 0, 235, 255, 0, 238, 255, 0, 241, 255, 0, 244, 255, 0, 247, 255, 0, 250, 255, 0, 253, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 23, 255, 0, 46, 255, 0, 70, 255, 0, 93, 255, 0, 116, 255, 0, 139, 255, 0, 163, 255, 0, 186, 255, 0, 209, 255, 0, 232, 255, 0, 255, 255, 0, 255, 237, 0, 255, 218, 0, 255, 200, 0, 255, 181, 0, 255, 163, 0, 255, 146, 0, 255, 128, 0, 255, 111, 0, 255, 93, 0, 255, 76, 0, 255, 63, 0, 255, 51, 0, 255, 42, 0, 255, 39, 0, 255, 36, 0, 255, 33, 0, 255, 30, 0, 255, 27, 0, 255, 24, 0, 255, 21, 0, 255, 18, 0, 255, 15, 0, 255, 12, 0, 255, 9, 0, 255, 6, 0, 255, 3, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0 };

int cc_[300] = { 0, 0, 0, 0, 0, 0, 0, 0, 255, 0, 60, 255, 0, 100, 255, 0, 125, 255, 0, 140, 255, 0, 155, 255, 0, 170, 255, 0, 180, 255, 0, 190, 255, 0, 200, 255, 0, 210, 255, 0, 220, 255, 0, 225, 255, 0, 230, 255, 0, 235, 255, 0, 238, 255, 0, 241, 255, 0, 244, 255, 0, 247, 255, 0, 250, 255, 0, 253, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 255, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 23, 255, 0, 46, 255, 0, 70, 255, 0, 93, 255, 0, 116, 255, 0, 139, 255, 0, 163, 255, 0, 186, 255, 0, 209, 255, 0, 232, 255, 0, 255, 255, 0, 255, 237, 0, 255, 218, 0, 255, 200, 0, 255, 181, 0, 255, 163, 0, 255, 146, 0, 255, 128, 0, 255, 111, 0, 255, 93, 0, 255, 76, 0, 255, 63, 0, 255, 51, 0, 255, 42, 0, 255, 39, 0, 255, 36, 0, 255, 33, 0, 255, 30, 0, 255, 27, 0, 255, 24, 0, 255, 21, 0, 255, 18, 0, 255, 15, 0, 255, 12, 0, 255, 9, 0, 255, 6, 0, 255, 3, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0, 255, 0, 0 };


__device__
real b_coeff[20] = {
    1.0000000000000000000000000000000,
    0.0833333333333333333333333333333,
    -0.0013888888888888888888888888888,
    3.3068783068783068783068783068783E-5,
    -8.2671957671957671957671957671958E-7,
    2.0876756987868098979210090321201E-8,
    -5.2841901386874931848476822021796E-10,
    1.3382536530684678832826980975129E-11,
    -3.3896802963225828668301953912494E-13,
    8.5860620562778445641359054504256E-15,
    -2.1748686985580618730415164238659E-16,
    5.5090028283602295152026526089023E-18,
    -1.3954464685812523340707686264064E-19,
    3.5347070396294674716932299778038E-21,
    -8.9535174266605480875210207537274E-23,
    2.2679524523376830603109507388682E-24
    -5.7447906688722024452638819876070E-26,
    1.4551724756148649018662648672713E-27,
    -3.6859949406653101781817824799086E-29,
    9.3367342570950446720325551527856E-31
};

__device__
real g_coeff[15] = {
    0.99999999999999709182,
    57.15623566586292351700,
    -59.59796035547549124800,
    14.13609797474174717400,
    -0.491913816097620199780,
    0.33994649984811888699E-4,
    0.46523628927048575665E-4,
    -0.98374475304879564677E-4,
    0.15808870322491248884E-3,
    -0.21026444172410488319E-3,
    0.21743961811521264320E-3,
    -0.16431810653676389022E-3,
    0.84418223983852743293E-4,
    -0.26190838401581408670E-4,
    0.36899182659531622704E-5
};

std::ostream& operator<<(std::ostream &out, const complex &z) {
    out << z.re << " + " << z.im << "i";
    return out;
}
std::ostream& operator<<(std::ostream &out, const Settings &settings) {
    out << "[" << settings.min << ", " << settings.max << "], " << settings.res << "ppu";
    return out;
}

__device__
void print_complex(const char* label, int i, complex z) {
    printf("%s %d %.16lG + %.16lGi\n", label, i, z.re, z.im);
}

__device__
real complex::arg() const {
    if (im != 0.0) return 2.0 * atan((mod() - re) / im);
    if (re > 0.0) return 0.0;
    if (re < 0.0) return CUDART_PI;
    return CUDART_NAN;
}

__device__
real complex::mod() const {
    return sqrt((re * re) + (im * im));
}

__device__
complex complex::ccos(complex const& c) {
    return complex(
        cos(c.re) * cosh(c.im),
        -1.0 * sin(c.re) * sinh(c.im)
    );
}

__device__
complex complex::cexp(complex const& c) {
    return complex(
        exp(c.re) * cos(c.im),
        exp(c.re) * sin(c.im)
    );
}

__device__
complex complex::clog(complex const& c) {
    return complex(
        log(c.mod()),
        c.arg()
    );
}

__device__
complex complex::cpow(real x, complex const& exp) {
    return complex(
        pow(x, exp.re) * cos((exp.im) * log(x)),
        pow(x, exp.re) * sin((exp.im) * log(x))
    );
}

__device__
complex complex::cpow(complex const& x, complex const& exp) {
    return cexp(exp * clog(x));
}

__device__
complex zeta(complex s) {
    complex z = 0, g = 1;
    if (s.re < 0) {
        if (abs(s.im) < MAX_GAMMA) {
            s = complex(1) - s;
            g = complex_gamma(s);
            z = ems(s);
            z = z * g * complex(2);
            z = z * complex::cpow(TWO_PI, -s);
            z = z * complex::ccos(complex(CUDART_PIO2) * s);
        }
        else {
            z = ems(s);
        }
    }
    else {
        z = ems(s);
    }
    return z;
}

__device__
complex ems(complex s) {
    int N = (int) s.mod(), k;
    complex z = 0, t = 0, temp = 0;
    if (N > MAX_N) N = MAX_N;
    if (N < MIN_N) N = MIN_N;
    for (k = 1; k < N; k++) {
        z = z + complex::cpow(k, -s);
    }
    // print_complex("ems", 0, z);
    z = z + (complex::cpow(N, complex(1) - s) / (s - 1));
    // print_complex("ems", 1, z);
    // z = fp_add(z, fp_div(fp_i_c_pow(N, fp_add(complex(1.0, 0.0), fp_neg(s))), fp_add(s, complex(-1.0, 0.0))));
    z = z + complex(0.5) * complex::cpow(N, -s);
    // print_complex("ems", 2, z);
    // z = fp_add(z, fp_mul(complex(0.5, 0.0), fp_i_c_pow(N, fp_neg(s))));
    for (k = 1; k < 20; k++) {
        t = t + complex(b_coeff[k]) * pochhammer(s, (2 * k) - 1) * complex::cpow(N, complex(1) - s - (2 * k));
        // t = fp_add(t, fp_mul(complex(b_coeff[k], 0.0), fp_mul(fp_pochhammer(s, (2 * k) - 1), fp_i_c_pow(N, fp_add(complex(1 - (2 * k), 0.0), fp_neg(s))))));
        if ((t - temp).re == 0) {
            break;
        }
        // if (fp_add(t, fp_neg(temp)).re == 0.0) break;
        temp = t;
    }
    // print_complex("ems", 3, z);
    return z + t;
}

__device__
complex pochhammer(complex s, int n) {
    int i;
    complex poch_val = 1;
    for (i = 0; i < n; i++) {
        poch_val = poch_val * (s + i);
        // poch_val = fp_mul(poch_val, fp_add(s, complex(i, 0.0)));
    }
    return poch_val;
}

__device__
complex complex_gamma(complex s) {
    int i;
    complex g = g_coeff[0];
    s = s - 1;
    for (i = 1; i < 15; i++) {
        g = g + (complex(g_coeff[i]) / (s + i));
        // g = fp_add(g, fp_div(complex(g_coeff[i], 0.0), fp_add(s, complex(i, 0.0))));
    }
    g = g * complex(SQRT_TWO_PI);
    // g = fp_mul(g, complex(SQRT_TWO_PI, 0.0));
    g = g * complex::cpow(s + complex(5.2421875), s + complex(0.5));
    // g = fp_mul(g, fp_c_c_pow(fp_add(s, complex(5.2421875, 0.0)), fp_add(s, complex(0.5, 0.0))));
    g = g * complex::cexp(complex(-5.2421875) - s);
    // g = fp_mul(g, fp_exp(fp_add(complex(-5.2421875, 0.0), fp_neg(s))));
    return g;
}

__device__
int iterate(complex s) {
    int i = 0;
    real cabs_z = 0.0, diff = 100;
    complex z = 0;
    while (!isnan(cabs_z) && diff > EPSILON && cabs_z < CABS_Z_MAX && i < MAX_ITS) {
        z = zeta(s);
        // print_complex("iter", i, z);
        diff = (z - s).mod();
        // diff = abs(z.re - s.re);
        cabs_z = z.mod();
        i++;
        s = z;
    }
    if (!isnan(cabs_z)) {
        if (cabs_z >= CABS_Z_MAX) {
            if (z.re < 0.0) {
                i += 1;
            } else {
                i += 1;
                i += 1;
            }
        }
    }
    return i;
}

__global__
void zeta_kernel(Settings settings, int width, int height, byte *pixels) {
    int start = blockIdx.x * blockDim.x + threadIdx.x;
    int stride = blockDim.x * gridDim.x;
    for (int i = start; i < width * height; i += stride) {
        int x = i % width;
        int y = i / width;
        real u = x / (real) width;
        real v = y / (real) height;
        complex min = settings.min;
        complex max = settings.max;
        complex range = max - min;
        complex s = min + complex(range.re * u, range.im * v);
        int iterations = iterate(s);
        pixels[(i * 3) + 0] = cc[(iterations * 3) + 0];
        pixels[(i * 3) + 1] = cc[(iterations * 3) + 1];
        pixels[(i * 3) + 2] = cc[(iterations * 3) + 2];
        // pixels[(i * 3) + 0] = u * 255;
        // pixels[(i * 3) + 1] = v * 255;
        // pixels[(i * 3) + 2] = 128;
    }
}

__global__
void calc_iterations(complex min, complex max, uint size, uint *data) {
    int start = blockIdx.x * blockDim.x + threadIdx.x;
    int stride = blockDim.x * gridDim.x;
    for (int i = start; i < size*size; i += stride) {
        int x = i % size;
        int y = i / size;
        real u = x / (real) size;
        real v = y / (real) size;
        complex range = max - min;
        complex s = min + complex(range.re * u, range.im * v);
        int iterations = iterate(s);
        data[i] = iterations;
    }
}

__global__
void zeta_kernel_lut(Settings settings, int width, int height, byte *pixels, ZetaLUTCollection *luts) {
    int start = blockIdx.x * blockDim.x + threadIdx.x;
    int stride = blockDim.x * gridDim.x;
    for (int i = start; i < width * height; i += stride) {
        int x = i % width;
        int y = i / width;
        real u = x / (real) width;
        real v = y / (real) height;
        complex min = settings.min;
        complex max = settings.max;
        complex range = max - min;
        complex s = min + complex(range.re * u, range.im * v);
        rgb_t color = {0, 0, 0};
        if (!luts->lookup(zeta(s), &color)) {
            color = {0, 0, 0};
        }
        pixels[(i * 3) + 0] = color.red;
        pixels[(i * 3) + 1] = color.green;
        pixels[(i * 3) + 2] = color.blue;
    }
}

__global__
void zeta_kernel_single(complex z) {
    iterate(z);
}

__host__
void write_image(Settings settings) {
    int width = (int) ((settings.max.re - settings.min.re) * settings.res);
    int height = (int) ((settings.max.im - settings.min.im) * settings.res);
    int procCount;
    int devId;
    cudaGetDevice(&devId);
    cudaDeviceGetAttribute(&procCount, cudaDevAttrMultiProcessorCount, devId);
    byte *pixels;
    cudaMallocManaged(&pixels, width * height * 3);
    //zeta_kernel<<<1, 1>>>(settings, width, height, pixels);
    zeta_kernel<<<procCount * 32, 256>>>(settings, width, height, pixels);
    cudaDeviceSynchronize();
    bitmap_image image(width, height);
    image.clear();
    for (int i = 0; i < width; i ++) {
        for (int j = 0; j < height; j ++) {
            int pixel = (j * width + i) * 3;
            rgb_t color = {
                pixels[pixel + 0],
                pixels[pixel + 1],
                pixels[pixel + 2]
            };
            image.set_pixel(i, height - j - 1, color);
        }
    }
    cudaFree(pixels);
    image.save_image("zeta.bmp");
}

__host__
void write_image_(int width, int height, uint* data) {
    unsigned char iterations;
    bitmap_image image(width, height);
    image.clear();
    for (int i = 0; i < width; i ++) {
        for (int j = 0; j < height; j ++) {
            int pixel = (j * width + i);

            iterations = (unsigned char)data[pixel];

            rgb_t color = {
                (unsigned char)cc_[(iterations * 3) + 0],
                (unsigned char)cc_[(iterations * 3) + 1],
                (unsigned char)cc_[(iterations * 3) + 2]
            };

            image.set_pixel(i, height - j - 1, color);
        }
    }
    image.save_image("zeta.bmp");
}

__host__
void write_image_lut(Settings settings, const ZetaLUTCollection& luts) {
    int width = (int) ((settings.max.re - settings.min.re) * settings.res);
    int height = (int) ((settings.max.im - settings.min.im) * settings.res);
    int procCount;
    int devId;
    cudaGetDevice(&devId);
    cudaDeviceGetAttribute(&procCount, cudaDevAttrMultiProcessorCount, devId);
    byte *pixels;
    cudaMallocManaged(&pixels, width * height * 3);
    ZetaLUTCollection *sharedluts;
    cudaMallocManaged(&sharedluts, sizeof(ZetaLUTCollection));
    *sharedluts = luts;
    //zeta_kernel<<<1, 1>>>(settings, width, height, pixels);
    zeta_kernel_lut<<<procCount * 32, 256>>>(settings, width, height, pixels, sharedluts);
    cudaDeviceSynchronize();
    bitmap_image image(width, height);
    image.clear();
    for (int i = 0; i < width; i ++) {
        for (int j = 0; j < height; j ++) {
            int pixel = (j * width + i) * 3;
            rgb_t color = {
                pixels[pixel + 0],
                pixels[pixel + 1],
                pixels[pixel + 2]
            };
            image.set_pixel(i, height - j - 1, color);
        }
    }
    cudaFree(pixels);
    cudaFree(sharedluts);
    image.save_image("zeta.bmp");
}

__host__
void iterate_single(complex z) {
    zeta_kernel_single<<<1, 1>>>(z);
    cudaDeviceSynchronize();
}

// void fp_image(real rl, real rh, real il, real ih, int res)
// {
//     char f_name[256];
//     unsigned char r = 0, g = 0, b = 0;
//     int its, row, pixel, width = ((rh - rl) * res) + 1, height = ((ih - il) * res) + 1, row_bytes = get_row_bytes(width);
//     real EPSILON = 1.0 / pow(10, 15), re, im, d = 1.0 / (real)res;
//     if (rh <= rl || ih <= il || res < 1) return;
//     sprintf(f_name, "RZ %.12lG, %.12lG, %.12lG, %.12lG, %d.bmp", rl, rh, il, ih, res);
//     printf("Initialising bitmap image...");
//     FILE *bmp1 = get_bmp(f_name, width, height, row_bytes, 0);
//     if (!bmp1) return;
//     printf("done\n\n");
//     im = il;
//     for (row = 1; row <= height; row++) {
//         re = rl;
//         for (pixel = 1; pixel <= width; pixel++) {
//             its = fp_iterate(complex(re, im), EPSILON, 0);
//             r = cc[(its * 3) + 0];
//             g = cc[(its * 3) + 1];
//             b = cc[(its * 3) + 2];
//             if (write_pixel(bmp1, row, pixel, row_bytes, r, g, b) != 3) return;
//             re += d;
//         }
//         im += d;
//         printf("row %d of %d\r", row, height);
//     }
//     fclose(bmp1);
//     printf("\n\n");
// }