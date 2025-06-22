#include <stdint.h>
#include <string.h>

#ifdef __AVX2__
#include <immintrin.h>
#elif defined(__ARM_NEON)
#include <arm_neon.h>
#endif

// Calculate MSE for RGBA images using SIMD
uint64_t compute_mse_rgba_simd(const uint8_t* pix1, const uint8_t* pix2, int length, int has_alpha) {
    uint64_t sum = 0;
    int i = 0;
    
#ifdef __AVX2__
    // AVX2: Process 32 bytes (8 pixels) at a time
    __m256i sum_vec = _mm256_setzero_si256();
    
    for (; i + 32 <= length; i += 32) {
        __m256i p1 = _mm256_loadu_si256((const __m256i*)(pix1 + i));
        __m256i p2 = _mm256_loadu_si256((const __m256i*)(pix2 + i));
        
        // Unpack to 16-bit to avoid overflow
        __m256i p1_lo = _mm256_unpacklo_epi8(p1, _mm256_setzero_si256());
        __m256i p1_hi = _mm256_unpackhi_epi8(p1, _mm256_setzero_si256());
        __m256i p2_lo = _mm256_unpacklo_epi8(p2, _mm256_setzero_si256());
        __m256i p2_hi = _mm256_unpackhi_epi8(p2, _mm256_setzero_si256());
        
        // Calculate differences
        __m256i diff_lo = _mm256_sub_epi16(p1_lo, p2_lo);
        __m256i diff_hi = _mm256_sub_epi16(p1_hi, p2_hi);
        
        // Square differences
        __m256i sq_lo = _mm256_mullo_epi16(diff_lo, diff_lo);
        __m256i sq_hi = _mm256_mullo_epi16(diff_hi, diff_hi);
        
        // If not using alpha, mask out every 4th component
        if (!has_alpha) {
            const __m256i mask = _mm256_set_epi16(
                0, -1, -1, -1, 0, -1, -1, -1,
                0, -1, -1, -1, 0, -1, -1, -1
            );
            sq_lo = _mm256_and_si256(sq_lo, mask);
            sq_hi = _mm256_and_si256(sq_hi, mask);
        }
        
        // Accumulate to 32-bit
        __m256i sum_lo = _mm256_unpacklo_epi16(sq_lo, _mm256_setzero_si256());
        __m256i sum_hi = _mm256_unpackhi_epi16(sq_lo, _mm256_setzero_si256());
        sum_vec = _mm256_add_epi32(sum_vec, sum_lo);
        sum_vec = _mm256_add_epi32(sum_vec, sum_hi);
        
        sum_lo = _mm256_unpacklo_epi16(sq_hi, _mm256_setzero_si256());
        sum_hi = _mm256_unpackhi_epi16(sq_hi, _mm256_setzero_si256());
        sum_vec = _mm256_add_epi32(sum_vec, sum_lo);
        sum_vec = _mm256_add_epi32(sum_vec, sum_hi);
    }
    
    // Horizontal sum
    __m128i sum_128 = _mm_add_epi32(
        _mm256_castsi256_si128(sum_vec),
        _mm256_extracti128_si256(sum_vec, 1)
    );
    sum_128 = _mm_hadd_epi32(sum_128, sum_128);
    sum_128 = _mm_hadd_epi32(sum_128, sum_128);
    sum += _mm_cvtsi128_si32(sum_128);
    
#elif defined(__ARM_NEON)
    // NEON: Process 16 bytes (4 pixels) at a time
    uint32x4_t sum_vec = vdupq_n_u32(0);
    
    for (; i + 16 <= length; i += 16) {
        uint8x16_t p1 = vld1q_u8(pix1 + i);
        uint8x16_t p2 = vld1q_u8(pix2 + i);
        
        // Convert to 16-bit and calculate differences
        int16x8_t diff_lo = vreinterpretq_s16_u16(vsubl_u8(vget_low_u8(p1), vget_low_u8(p2)));
        int16x8_t diff_hi = vreinterpretq_s16_u16(vsubl_u8(vget_high_u8(p1), vget_high_u8(p2)));
        
        // Square differences
        int16x8_t sq_lo = vmulq_s16(diff_lo, diff_lo);
        int16x8_t sq_hi = vmulq_s16(diff_hi, diff_hi);
        
        // If not using alpha, mask out every 4th component
        if (!has_alpha) {
            const int16x8_t mask_lo = {-1, -1, -1, 0, -1, -1, -1, 0};
            const int16x8_t mask_hi = {-1, -1, -1, 0, -1, -1, -1, 0};
            sq_lo = vandq_s16(sq_lo, mask_lo);
            sq_hi = vandq_s16(sq_hi, mask_hi);
        }
        
        // Accumulate to 32-bit
        sum_vec = vaddq_u32(sum_vec, vaddl_u16(vget_low_u16(vreinterpretq_u16_s16(sq_lo)), 
                                                vget_high_u16(vreinterpretq_u16_s16(sq_lo))));
        sum_vec = vaddq_u32(sum_vec, vaddl_u16(vget_low_u16(vreinterpretq_u16_s16(sq_hi)), 
                                                vget_high_u16(vreinterpretq_u16_s16(sq_hi))));
    }
    
    // Horizontal sum
    uint32x2_t sum_pair = vadd_u32(vget_low_u32(sum_vec), vget_high_u32(sum_vec));
    sum += vget_lane_u32(vpadd_u32(sum_pair, sum_pair), 0);
#endif
    
    // Process remaining pixels
    for (; i < length; i += 4) {
        int32_t diffR = (int32_t)pix1[i] - (int32_t)pix2[i];
        int32_t diffG = (int32_t)pix1[i+1] - (int32_t)pix2[i+1];
        int32_t diffB = (int32_t)pix1[i+2] - (int32_t)pix2[i+2];
        
        sum += (uint64_t)(diffR * diffR) + (uint64_t)(diffG * diffG) + (uint64_t)(diffB * diffB);
        
        if (has_alpha) {
            int32_t diffA = (int32_t)pix1[i+3] - (int32_t)pix2[i+3];
            sum += (uint64_t)(diffA * diffA);
        }
    }
    
    return sum;
}

// Calculate MSE for YCbCr images using SIMD
uint64_t compute_mse_ycbcr_simd(const uint8_t* y1, const uint8_t* y2, int y_len,
                                const uint8_t* cb1, const uint8_t* cb2, int cb_len,
                                const uint8_t* cr1, const uint8_t* cr2, int cr_len) {
    uint64_t sum = 0;
    int i = 0;
    
#ifdef __AVX2__
    // Process Y channel with AVX2
    __m256i sum_vec = _mm256_setzero_si256();
    
    for (; i + 32 <= y_len; i += 32) {
        __m256i p1 = _mm256_loadu_si256((const __m256i*)(y1 + i));
        __m256i p2 = _mm256_loadu_si256((const __m256i*)(y2 + i));
        
        // Calculate absolute differences
        __m256i diff = _mm256_sub_epi8(_mm256_max_epu8(p1, p2), _mm256_min_epu8(p1, p2));
        
        // Square differences (need to unpack to 16-bit first)
        __m256i diff_lo = _mm256_unpacklo_epi8(diff, _mm256_setzero_si256());
        __m256i diff_hi = _mm256_unpackhi_epi8(diff, _mm256_setzero_si256());
        
        __m256i sq_lo = _mm256_mullo_epi16(diff_lo, diff_lo);
        __m256i sq_hi = _mm256_mullo_epi16(diff_hi, diff_hi);
        
        // Accumulate to 32-bit
        sum_vec = _mm256_add_epi32(sum_vec, _mm256_unpacklo_epi16(sq_lo, _mm256_setzero_si256()));
        sum_vec = _mm256_add_epi32(sum_vec, _mm256_unpackhi_epi16(sq_lo, _mm256_setzero_si256()));
        sum_vec = _mm256_add_epi32(sum_vec, _mm256_unpacklo_epi16(sq_hi, _mm256_setzero_si256()));
        sum_vec = _mm256_add_epi32(sum_vec, _mm256_unpackhi_epi16(sq_hi, _mm256_setzero_si256()));
    }
    
    // Horizontal sum
    __m128i sum_128 = _mm_add_epi32(
        _mm256_castsi256_si128(sum_vec),
        _mm256_extracti128_si256(sum_vec, 1)
    );
    sum_128 = _mm_hadd_epi32(sum_128, sum_128);
    sum_128 = _mm_hadd_epi32(sum_128, sum_128);
    sum += _mm_cvtsi128_si32(sum_128);
    
#elif defined(__ARM_NEON)
    // Process Y channel with NEON
    uint32x4_t sum_vec = vdupq_n_u32(0);
    
    for (; i + 16 <= y_len; i += 16) {
        uint8x16_t p1 = vld1q_u8(y1 + i);
        uint8x16_t p2 = vld1q_u8(y2 + i);
        
        // Calculate differences
        int16x8_t diff_lo = vreinterpretq_s16_u16(vsubl_u8(vget_low_u8(p1), vget_low_u8(p2)));
        int16x8_t diff_hi = vreinterpretq_s16_u16(vsubl_u8(vget_high_u8(p1), vget_high_u8(p2)));
        
        // Square differences
        int16x8_t sq_lo = vmulq_s16(diff_lo, diff_lo);
        int16x8_t sq_hi = vmulq_s16(diff_hi, diff_hi);
        
        // Accumulate
        sum_vec = vaddq_u32(sum_vec, vaddl_u16(vget_low_u16(vreinterpretq_u16_s16(sq_lo)), 
                                                vget_high_u16(vreinterpretq_u16_s16(sq_lo))));
        sum_vec = vaddq_u32(sum_vec, vaddl_u16(vget_low_u16(vreinterpretq_u16_s16(sq_hi)), 
                                                vget_high_u16(vreinterpretq_u16_s16(sq_hi))));
    }
    
    // Horizontal sum
    uint32x2_t sum_pair = vadd_u32(vget_low_u32(sum_vec), vget_high_u32(sum_vec));
    sum += vget_lane_u32(vpadd_u32(sum_pair, sum_pair), 0);
#endif
    
    // Process remaining Y pixels
    for (; i < y_len; i++) {
        int32_t diff = (int32_t)y1[i] - (int32_t)y2[i];
        sum += (uint64_t)(diff * diff);
    }
    
    // Process Cb and Cr channels (usually smaller, so no SIMD)
    for (i = 0; i < cb_len; i++) {
        int32_t diff = (int32_t)cb1[i] - (int32_t)cb2[i];
        sum += (uint64_t)(diff * diff);
    }
    
    for (i = 0; i < cr_len; i++) {
        int32_t diff = (int32_t)cr1[i] - (int32_t)cr2[i];
        sum += (uint64_t)(diff * diff);
    }
    
    return sum;
}