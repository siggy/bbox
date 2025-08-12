/*
  SCORPIO + Adafruit_NeoPXL8 framed USB control
  - 8 strands, 144 pixels each, GRBW
  - USB CDC protocol:
      [0xAA][len_hi][len_lo][payload ...][chk]
      payload is N * 6 bytes: [si, pi, r, g, b, w]
      chk = XOR of all payload bytes
*/

#include <Adafruit_NeoPXL8.h>
#include <Adafruit_NeoPixel.h>
#include <vector>
#include <math.h>

// -------- Configuration --------
#define NUM_STRANDS 2
#define STRAND_LEN 240
#define COLOR_ORDER NEO_GRBW

// SCORPIO RP2040 PIO pins (GP16..GP17)
int8_t pins[8] = { 16, 17, -1, -1, -1, -1, -1, -1 };

// Onboard heartbeat NeoPixel
#ifndef PIN_NEOPIXEL
#define PIN_NEOPIXEL 16 // fallback, most RP2040 boards define this
#endif

// -------- Globals --------
Adafruit_NeoPXL8 leds(STRAND_LEN, pins, COLOR_ORDER);
Adafruit_NeoPixel hb(1, PIN_NEOPIXEL, NEO_GRB + NEO_KHZ800);

std::vector<uint8_t> buf; // rolling USB buffer

// Protocol constants
const uint8_t START = 0xAA;

// -------- Helpers --------
static inline uint32_t packGRBW(uint8_t r, uint8_t g, uint8_t b, uint8_t w)
{
  // Adafruit_NeoPXL8 uses the same Color() signature as NeoPixel (RGB or RGBW).
  return leds.Color(r, g, b, w);
}

void heartbeat()
{
  // Simple green<->blue breathing at ~1 Hz
  static uint32_t t0 = millis();
  float t = (millis() - t0) / 1000.0f;
  float pulse = (sinf(2.0f * M_PI * t) + 1.0f) * 0.5f; // 0..1
  uint8_t g = (uint8_t)(pulse * 64.0f);
  uint8_t b = (uint8_t)((1.0f - pulse) * 64.0f);
  hb.setPixelColor(0, hb.Color(0, g, b));
  hb.show();
}

bool parseAndApplyFrames()
{
  // Returns true if any pixels were changed (to optionally optimize show()).
  bool changed = false;

  // We may have multiple frames in the buffer.
  for (;;)
  {
    if (buf.size() < 4)
      break; // need at least start + len + chk
    if (buf[0] != START)
    {
      buf.erase(buf.begin());
      continue;
    }

    uint16_t length = ((uint16_t)buf[1] << 8) | buf[2];
    uint32_t total = 3u + length + 1u; // start + len(2) + payload + chk
    if (buf.size() < total)
      break; // wait for full frame

    // Verify checksum (XOR of payload)
    uint8_t x = 0;
    for (uint32_t i = 0; i < length; i++)
      x ^= buf[3 + i];
    uint8_t chk = buf[3 + length];

    if (x == chk && (length % 6u) == 0u)
    {
      // Apply pixels
      for (uint32_t i = 3; i < 3u + length; i += 6)
      {
        uint8_t si = buf[i + 0];
        uint8_t pi = buf[i + 1];
        uint8_t r = buf[i + 2];
        uint8_t g = buf[i + 3];
        uint8_t b = buf[i + 4];
        uint8_t w = buf[i + 5];

        if (si < NUM_STRANDS && pi < STRAND_LEN)
        {
          uint32_t idx = (uint32_t)si * STRAND_LEN + pi;
          leds.setPixelColor(idx, packGRBW(r, g, b, w));
          changed = true;
        }
      }
    }
    // Drop processed frame (valid or not — resync happens via START search)
    buf.erase(buf.begin(), buf.begin() + total);
  }
  return changed;
}

// -------- Arduino lifecycle --------
void setup()
{
  // USB serial
  Serial.begin(115200);
  // Give USB a moment on cold-boot; non-blocking if already up.
  uint32_t tstart = millis();
  while (!Serial && (millis() - tstart < 1500))
  {
  }

  // Heartbeat LED
  hb.begin();
  hb.setBrightness(128);
  hb.setPixelColor(0, hb.Color(0, 255, 0)); // green on boot
  hb.show();

  // NeoPXL8 start
  if (!leds.begin())
  {
    pinMode(LED_BUILTIN, OUTPUT);
    for (;;)
      digitalWrite(LED_BUILTIN, (millis() / 250) & 1); // fatal blink
  }
  leds.setBrightness(255); // full; adjust if needed

  // Startup: all red (R channel) for 1 second, then clear
  leds.fill(packGRBW(255, 0, 0, 0));
  leds.show();
  delay(2000);
  leds.clear();
  leds.show();

  buf.reserve(4096);
}

void loop()
{
  heartbeat();

  // Pull any available USB bytes into rolling buffer
  while (Serial.available() > 0)
  {
    buf.push_back((uint8_t)Serial.read());
    // Prevent unbounded growth on junk input
    if (buf.size() > 65536)
      buf.erase(buf.begin(), buf.begin() + 32768);
  }

  bool changed = parseAndApplyFrames();

  if (changed)
  {
    leds.show();
  }

  // Tiny pause keeps loop snappy without starving USB
  // (lower CPU, but still very responsive)
  delay(1);
}
