#include <Adafruit_NeoPixel.h>
#include <math.h>

#define LED_COUNT     353
#define LED_PIN       6
#define COMET_LENGTH  25
#define SPACING       50
#define DELAY_MS      50

// Use bitwise OR to combine flags
Adafruit_NeoPixel strip(LED_COUNT, LED_PIN, NEO_GRB | NEO_KHZ800);

uint16_t head = 0;

// Background (turquoise) and comet‐center (yellow) RGB components
const uint8_t baseR = 102, baseG = 255, baseB = 255;  // turquoise
const uint8_t peakR = 255, peakG = 255, peakB =  51;  // yellow

// Exponent > 1 gives a sharper falloff toward the ends
const float EXPONENT = 2.0;

void setup() {
  strip.begin();
  strip.setBrightness(16);
  strip.show();  // all off
}

void loop() {
  // 1) Fill entire strip with turquoise
  for(uint16_t i = 0; i < LED_COUNT; i++) {
    strip.setPixelColor(i, baseR, baseG, baseB);
  }

  // 2) Overlay multiple comets
  int stride     = COMET_LENGTH + SPACING;
  int numComets  = (LED_COUNT + stride - 1) / stride;  // ceil

  float mid = (COMET_LENGTH - 1) / 2.0;

  for(int k = 0; k < numComets; k++) {
    uint16_t cometHead = (head + k * stride) % LED_COUNT;
    for(uint16_t j = 0; j < COMET_LENGTH; j++) {
      // distance from center of comet
      float dist = fabs(j - mid);
      float norm = dist / mid;        // 0 at center → 1 at ends
      float t    = 1.0 - pow(norm, EXPONENT);  // 1 at center, 0 at ends

      // interpolate each channel
      uint8_t r = baseR + (uint8_t)((peakR - baseR) * t + 0.5);
      uint8_t g = baseG + (uint8_t)((peakG - baseG) * t + 0.5);
      uint8_t b = baseB + (uint8_t)((peakB - baseB) * t + 0.5);

      // wrap around strip
      uint16_t idx = (cometHead + LED_COUNT - j) % LED_COUNT;
      strip.setPixelColor(idx, r, g, b);
    }
  }

  strip.show();

  // advance head
  head = (head + 1) % LED_COUNT;
  delay(DELAY_MS);
}
