#include <Adafruit_NeoPixel.h>
#include <math.h>

#define LED_COUNT        353
#define LED_PIN          6
#define DELAY_MS         20
#define SEGMENT_COUNT    40      // how many simultaneous flicker segments
#define MIN_SEG_LEN      5       // smallest segment length
#define MAX_SEG_LEN      30      // largest segment length
#define FLICKER_CYCLE_MS 100    // full dim→bright→dim cycle
#define SHAPE_EXP        2.0     // power curve exponent for segment profile

// strip uses GRB ordering on WS2812B
Adafruit_NeoPixel strip(LED_COUNT, LED_PIN, NEO_GRB | NEO_KHZ800);

// describes one flicker segment
struct Segment {
  uint16_t center;  // center pixel index
  uint8_t  length;  // total length in pixels
};
Segment segments[SEGMENT_COUNT];

// current phase [0..1) of the flicker cycle
float flickerPhase = 0;

void randomizeSegments() {
  for(int i = 0; i < SEGMENT_COUNT; i++) {
    segments[i].center = random(LED_COUNT);
    segments[i].length = random(MIN_SEG_LEN, MAX_SEG_LEN + 1);
  }
}

void setup() {
  strip.begin();
  strip.setBrightness(255);   // full‐bright background
  strip.show();               // clear
  randomSeed(micros());
  randomizeSegments();
}

void loop() {
  // 1) Fill entire strip with pure blue
  for(uint16_t i = 0; i < LED_COUNT; i++) {
    strip.setPixelColor(i, 0, 0, 255);
  }

  // 2) Compute current flicker intensity [0..1]
  flickerPhase += (float)DELAY_MS / FLICKER_CYCLE_MS;
  if(flickerPhase >= 1.0) {
    flickerPhase -= 1.0;
    randomizeSegments();  // new random segments each cycle
  }
  // sine‐based cycle gives smooth dim→bright→dim
  float phaseVal = sinf(flickerPhase * M_PI);

  // 3) Overlay each segment with a power‐curved dimming profile
  for(int s = 0; s < SEGMENT_COUNT; s++) {
    float half = (segments[s].length - 1) * 0.5;
    for(int j = 0; j < segments[s].length; j++) {
      float dist     = fabsf(j - half);
      float norm     = dist / half;             // 0 @ center → 1 @ ends
      float shape    = powf(1.0 - norm, SHAPE_EXP);
      float dimFrac  = shape * phaseVal;        // combine segment & flicker
      uint8_t b      = (uint8_t)(255 * (1.0 - dimFrac) + 0.5);
      int idx        = (segments[s].center + j - (int)half + LED_COUNT) % LED_COUNT;
      strip.setPixelColor(idx, 0, 0, b);
    }
  }

  strip.show();
  delay(DELAY_MS);
}
