#include <Adafruit_NeoPixel.h>
#include <math.h>

#define LED_COUNT        353
#define LED_PIN          6
#define DELAY_MS         10
#define SEGMENT_COUNT    40
#define MIN_SEG_LEN      5
#define MAX_SEG_LEN      50
#define FLICKER_CYCLE_MS 50
#define SHAPE_EXP        2.0f
#define LUT_SIZE         50
#define FLICKER_STEPS    (FLICKER_CYCLE_MS / DELAY_MS)  // e.g. 50/10 = 5
#define PI_F             3.14159265f

Adafruit_NeoPixel strip(LED_COUNT, LED_PIN, NEO_GRB | NEO_KHZ800);

struct Segment {
  uint16_t center;
  uint8_t  length;
  uint8_t  half;      // (length-1)/2
  uint8_t  lenMinus1; // length-1
};
static Segment segments[SEGMENT_COUNT];

// Precomputed tables
static float    shapeLUT[LUT_SIZE + 1];                   // powf(1 - x, exp)
static uint8_t  brightnessLUT[FLICKER_STEPS][LUT_SIZE + 1]; // final brightness per phase & shape index

static uint8_t flickerIdx = 0;

void randomizeSegments() {
  for(int i = 0; i < SEGMENT_COUNT; i++) {
    uint8_t L = random(MIN_SEG_LEN, MAX_SEG_LEN + 1);
    segments[i].center    = random(LED_COUNT);
    segments[i].length    = L;
    segments[i].lenMinus1 = L - 1;
    segments[i].half      = (L - 1) >> 1;
  }
}

void setup() {
  strip.begin();
  strip.setBrightness(255);
  strip.show();
  randomSeed(micros());

  // 1) Shape lookup: shapeLUT[i] = powf(1 - i/LUT_SIZE, SHAPE_EXP)
  for(int i = 0; i <= LUT_SIZE; i++) {
    float r = (float)i / LUT_SIZE;          // 0.0 → 1.0
    shapeLUT[i] = powf(1.0f - r, SHAPE_EXP);
  }

  // 2) Brightness lookup: for each flicker step and shape index
  for(int f = 0; f < FLICKER_STEPS; f++) {
    float phase = sinf(((float)f / FLICKER_STEPS) * PI_F);  // 0→1→0
    for(int s = 0; s <= LUT_SIZE; s++) {
      float b = 255.0f * (1.0f - shapeLUT[s] * phase);
      brightnessLUT[f][s] = (uint8_t)(b + 0.5f);
    }
  }

  randomizeSegments();
}

void loop() {
  // Advance flicker index and wrap
  if(++flickerIdx >= FLICKER_STEPS) {
    flickerIdx = 0;
    randomizeSegments();
  }

  // 1) Paint background bright blue
  for(uint16_t i = 0; i < LED_COUNT; i++) {
    strip.setPixelColor(i, 0, 0, 255);
  }

  // 2) Overlay segments using only integer math and table lookups
  for(int s = 0; s < SEGMENT_COUNT; s++) {
    uint8_t  len1 = segments[s].lenMinus1;
    uint8_t  half = segments[s].half;
    uint16_t cen  = segments[s].center;
    for(uint8_t j = 0; j <= len1; j++) {
      // compute |2*j - (length-1)|
      uint8_t diff = (j << 1) > len1 ? (j << 1) - len1 : len1 - (j << 1);
      // map diff to shape index
      uint8_t idx = (diff * LUT_SIZE + half) / len1;
      // lookup final brightness
      uint8_t b   = brightnessLUT[flickerIdx][idx];

      int pix = cen + j - half;
      if(pix < 0)          pix += LED_COUNT;
      else if(pix >= LED_COUNT) pix -= LED_COUNT;

      strip.setPixelColor(pix, 0, 0, b);
    }
  }

  strip.show();
  delay(DELAY_MS);
}
