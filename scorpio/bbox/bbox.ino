/*
  SCORPIO + Adafruit_NeoPXL8 framed USB control
  - 8 strands, 144 pixels each, GRBW
  - USB CDC protocol:
      [0xAA][len_hi][len_lo][payload ...][chk]
      payload is N * 6 bytes: [si, pi, r, g, b, w]
      chk = XOR of all payload bytes
*/

#include <Adafruit_NeoPXL8.h>
#include <vector>
#include <math.h>

// -------- Configuration --------
#define NUM_STRANDS 8
#define STRAND_LEN 144
#define COLOR_ORDER NEO_GRBW

// SCORPIO RP2040 PIO pins (GP16..GP23)
int8_t pins[8] = {16, 17, 18, 19, 20, 21, 22, 23};

// -------- Globals --------
Adafruit_NeoPXL8 leds(STRAND_LEN, pins, COLOR_ORDER);

std::vector<uint8_t> buf; // rolling USB buffer

// Protocol constants
const uint8_t START = 0xAA;

// add near the top (globals)
static uint32_t front_stamp_ms = 0;
static const uint32_t FRAME_STUCK_MS = 200;                             // resync window
static const uint32_t MAX_LEN = (uint32_t)NUM_STRANDS * STRAND_LEN * 6; // hard cap

// -------- Helpers --------
static inline uint32_t packGRBW(uint8_t r, uint8_t g, uint8_t b, uint8_t w)
{
  // Adafruit_NeoPXL8 uses the same Color() signature as NeoPixel (RGB or RGBW).
  return leds.Color(r, g, b, w);
}

// Heartbeat on built-in LED (no PIO/NeoPixel usage)
void heartbeat()
{
  static uint32_t t0 = millis();
  float t = (millis() - t0) / 1000.0f;
  float pulse = (sinf(2.0f * M_PI * t) + 1.0f) * 0.5f; // 0..1
  int duty = (int)(pulse * 255.0f);

  // Prefer PWM if available on LED_BUILTIN; otherwise fall back to binary
  analogWrite(LED_BUILTIN, duty);
}

bool parseAndApplyFrames()
{
  bool changed = false;

  for (;;)
  {
    if (buf.size() < 4)
      break; // need start + len_hi + len_lo + chk(min)

    // Resync until we find a START at front
    if (buf[0] != START)
    {
      buf.erase(buf.begin());
      front_stamp_ms = 0;
      continue;
    }

    // Have a START; if this is a new candidate frame, start a timer
    if (front_stamp_ms == 0)
      front_stamp_ms = millis();

    // read length
    uint16_t length = ((uint16_t)buf[1] << 8) | buf[2];

    // Sanity checks — drop the START and resync on bogus lengths
    if (length == 0 || length > MAX_LEN || (length % 6u) != 0u)
    {
      // bad frame header; drop START and try next byte
      buf.erase(buf.begin());
      front_stamp_ms = 0;
      continue;
    }

    uint32_t total = 3u + length + 1u; // start + len(2) + payload + chk
    if (buf.size() < total)
    {
      // Not enough data yet. If it's been too long, assume corrupted header and resync.
      if (millis() - front_stamp_ms > FRAME_STUCK_MS)
      {
        buf.erase(buf.begin()); // drop START; try to find next one
        front_stamp_ms = 0;
      }
      break; // wait for more bytes
    }

    // Verify checksum
    uint8_t x = 0;
    for (uint32_t i = 0; i < length; i++)
      x ^= buf[3 + i];
    uint8_t chk = buf[3 + length];

    if (x == chk)
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
          leds.setPixelColor(idx, leds.Color(r, g, b, w));
          changed = true;
        }
      }
    }
    // Drop processed (or bad‑checksum) frame and reset timer
    buf.erase(buf.begin(), buf.begin() + total);
    front_stamp_ms = 0;
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

  // Heartbeat LED (built-in)
  pinMode(LED_BUILTIN, OUTPUT);
  analogWrite(LED_BUILTIN, 128); // mid brightness on boot

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
