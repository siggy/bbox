/*
  Single-strand NeoPXL8 demo (pulses + twinkles)
  - Pulses: NO OVERLAP by position rule; spawn next after newest rear edge
            passes a random distance from the start (min MIN_PULSE_GAP_PX)
  - Smooth rendering: float accumulation + gamma + light temporal dithering
  - Twinkles: white channel, fixed rate; relocate each cycle, slow rise/fall
*/

#include <Adafruit_NeoPXL8.h>
#include <math.h>
#include <vector>

// ======== USER TUNABLES ========
#define STRAND_LEN 144
#define COLOR_ORDER NEO_GRBW
#define DATA_PIN 16

// Pulse motion
const float PULSE_SPEED = 4.0f; // px/sec

// Pulse size (HALF-length randomized); visible total ≈ 2*halfLen
const float PULSE_HALFLEN_MIN_PX = 1.0f;
const float PULSE_HALFLEN_MAX_PX = 7.0f;

// Spacing rule: spawn next pulse when the newest pulse's REAR edge (leftmost)
// has traveled at least: MIN_PULSE_GAP_PX + random(0..EXTRA_GAP_SPREAD_PX)
const float MIN_PULSE_GAP_PX = 1.0f;
const float EXTRA_GAP_SPREAD_PX = 18.0f; // randomness range; set 0 for fixed spacing

// Tail shaping — VERY SOFT EDGES
// Brightness = (0.5 * (1 + cos(pi * (d / (halfLen * EDGE_SOFTNESS))))) ^ GAMMA
const float EDGE_SOFTNESS = 2.0f;      // extend tails
const float PULSE_SHAPE_GAMMA = 0.35f; // <1 = gentler edges

// Temporal fade-in (attack) and fade-out (release)
const uint16_t ATTACK_MS = 1500;        // slow ramp-up
const uint16_t RELEASE_MS = 2000;       // slow fade-out at the end
const float PRE_ENTRY_MARGIN_PX = 6.0f; // spawn further off-strip

// Peak per-pulse (pre-gamma)
const uint8_t PULSE_PEAK_MAX = 160;

// Pulse color (same for all pulses)
const uint8_t PULSE_R = 255, PULSE_G = 0, PULSE_B = 0;

// Twinkles
const uint8_t TWINKLE_COUNT = 10;
const uint8_t TWINKLE_MAX_WHITE = 80;
const uint32_t TWINKLE_PERIOD_MS = 2400;

// ======== WIRING ========
int8_t pins[8] = {DATA_PIN, -1, -1, -1, -1, -1, -1, -1};
Adafruit_NeoPXL8 leds(STRAND_LEN, pins, COLOR_ORDER);

// ======== STATE ========
struct Pulse
{
  float center;     // center position (px)
  float halfLen;    // base half-length (px)
  float drawHalf;   // half-length used for drawing = halfLen * EDGE_SOFTNESS
  uint32_t bornMs;  // for attack envelope
  uint32_t endAtMs; // when rear edge has left the strip (for release window)
};

std::vector<Pulse> pulses;           // active pulses (dynamic)
float nextSpawnRearThreshold = 0.0f; // absolute rear-edge distance target from pixel 0

// float accumulators
float accR[STRAND_LEN], accG[STRAND_LEN], accB[STRAND_LEN], accW[STRAND_LEN];

// timing
uint32_t lastFrameMs = 0;
uint32_t frameCounter = 0; // for dithering phase

// --- Twinkles ---
struct Twinkle
{
  uint16_t pixel;
  int32_t phaseMs;
  float prevPhase;
};
Twinkle twinkles[TWINKLE_COUNT];

// ======== HELPERS ========
static inline uint32_t packColor(uint8_t r, uint8_t g, uint8_t b, uint8_t w)
{
  return leds.Color(r, g, b, w);
}
static inline float frand(float a, float b)
{
  return a + (b - a) * (float)random(0, 10000) / 10000.0f;
}
static inline float randGap()
{
  return MIN_PULSE_GAP_PX + frand(0.0f, EXTRA_GAP_SPREAD_PX);
}
// simple clamp 0..1
static inline float clamp01(float x) { return x < 0 ? 0 : (x > 1 ? 1 : x); }
// cosine ease 0..1 -> 0..1
static inline float easeCos(float u)
{
  if (u <= 0)
    return 0.0f;
  if (u >= 1)
    return 1.0f;
  return 0.5f * (1.0f - cosf((float)M_PI * u));
}
// pseudo-random in [0,1) (for dithering), takes full 32-bit frame counter
static inline float hash01(uint16_t x, uint32_t y)
{
  uint32_t h = 2166136261u;
  h = (h ^ x) * 16777619u;
  h = (h ^ (uint16_t)(y)) * 16777619u;
  h = (h ^ (uint16_t)(y >> 16)) * 16777619u;
  return (h & 0xFFFFFF) / 16777216.0f;
}

// Spawn a new pulse off-strip (to the left)
void spawnPulse()
{
  Pulse p;
  p.halfLen = frand(PULSE_HALFLEN_MIN_PX, PULSE_HALFLEN_MAX_PX);
  p.drawHalf = p.halfLen * EDGE_SOFTNESS;
  p.center = -(p.drawHalf + PRE_ENTRY_MARGIN_PX); // start very faint
  p.bornMs = millis();

  // travel time until rear edge exits the strip (center - drawHalf > STRAND_LEN)
  float travel = (STRAND_LEN + p.drawHalf + PRE_ENTRY_MARGIN_PX + 12.0f) / fmaxf(PULSE_SPEED, 1.0f);
  p.endAtMs = p.bornMs + (uint32_t)(travel * 1000.0f);

  pulses.push_back(p);

  // Set next rear-edge threshold (absolute distance from start)
  float newestRear = fmaxf(0.0f, p.center - p.drawHalf); // negative at spawn -> 0
  nextSpawnRearThreshold = newestRear + randGap();
}

void initTwinkles()
{
  for (uint8_t i = 0; i < TWINKLE_COUNT; i++)
  {
    twinkles[i].pixel = random(0, STRAND_LEN);
    twinkles[i].phaseMs = random(0, TWINKLE_PERIOD_MS);
    twinkles[i].prevPhase = 0.0f;
  }
}
static inline void relocateTwinkleToStart(Twinkle &t, uint32_t nowMs)
{
  t.pixel = random(0, STRAND_LEN);
  t.phaseMs = -(int32_t)(nowMs % TWINKLE_PERIOD_MS);
  t.prevPhase = 0.0f;
}

void renderFrame(float dt)
{
  // clear accumulators
  for (int i = 0; i < STRAND_LEN; i++)
    accR[i] = accG[i] = accB[i] = accW[i] = 0.0f;

  const float peakR = (PULSE_R / 255.0f) * (PULSE_PEAK_MAX / 255.0f);
  const float peakG = (PULSE_G / 255.0f) * (PULSE_PEAK_MAX / 255.0f);
  const float peakB = (PULSE_B / 255.0f) * (PULSE_PEAK_MAX / 255.0f);

  uint32_t nowMs = millis();

  // Advance & draw pulses; remove those that have fully passed
  for (size_t i = 0; i < pulses.size(); /* increment inside */)
  {
    Pulse &p = pulses[i];

    // move
    p.center += PULSE_SPEED * dt;

    // cull if done (rear edge beyond strip)
    if (nowMs > p.endAtMs || (p.center - p.drawHalf) > (STRAND_LEN + 1))
    {
      pulses.erase(pulses.begin() + i);
      continue;
    }

    // Temporal envelopes:
    // Attack  : 0→1 over ATTACK_MS from bornMs
    float attack = easeCos((float)(nowMs - p.bornMs) / (float)ATTACK_MS);

    // ✅ Correct release: 1→0 over the *last* RELEASE_MS before endAtMs
    int32_t tToEnd = (int32_t)p.endAtMs - (int32_t)nowMs; // ms until end
    float u = clamp01((float)tToEnd / (float)RELEASE_MS); // 1..0 over final window
    float release = easeCos(u);                           // high until close to the end, then eases down

    float env = attack * release;

    // Draw pulse with very soft spatial falloff
    int start = (int)floorf(p.center - p.drawHalf);
    int end = (int)ceilf(p.center + p.drawHalf);
    if (start < 0)
      start = 0;
    if (end > STRAND_LEN)
      end = STRAND_LEN;

    for (int px = start; px < end; px++)
    {
      float d = fabsf((float)px - p.center); // 0 at center → drawHalf at edges
      float x = d / p.drawHalf;              // 0..1
      if (x > 1.0f)
        continue;

      // Extremely soft raised-cosine with gamma < 1
      float bSpatial = powf(0.5f * (1.0f + cosf((float)M_PI * x)), PULSE_SHAPE_GAMMA);
      float b = env * bSpatial;

      accR[px] += peakR * b;
      accG[px] += peakG * b;
      accB[px] += peakB * b;
    }

    ++i;
  }

  // Spawn rule: first pulse immediately; then by rear-edge distance
  if (pulses.empty())
  {
    spawnPulse();
  }
  else
  {
    const Pulse &newest = pulses.back();
    float newestRear = newest.center - newest.drawHalf; // rear (left) edge
    float traveledFromStart = fmaxf(0.0f, newestRear);
    if (traveledFromStart >= nextSpawnRearThreshold)
    {
      spawnPulse();
    }
  }

  // Twinkles
  for (uint8_t i = 0; i < TWINKLE_COUNT; i++)
  {
    Twinkle &t = twinkles[i];
    float phase = (float)((int32_t)(nowMs + t.phaseMs) % (int32_t)TWINKLE_PERIOD_MS) / (float)TWINKLE_PERIOD_MS;
    if (t.prevPhase > 0.9f && phase < 0.1f)
    {
      relocateTwinkleToStart(t, nowMs);
      phase = 0.0f;
    }
    t.prevPhase = phase;

    float s = 0.5f * (1.0f - cosf(phase * 2.0f * (float)M_PI)); // 0..1
    float w = (s * TWINKLE_MAX_WHITE) / 255.0f;

    int px = t.pixel;
    if (px >= 0 && px < STRAND_LEN)
      accW[px] = fmaxf(accW[px], w);
  }

  // Quantize once, with gamma + light temporal dithering (long period)
  leds.clear();
  for (int px = 0; px < STRAND_LEN; px++)
  {
    float r = fminf(accR[px], 1.0f);
    float g = fminf(accG[px], 1.0f);
    float b = fminf(accB[px], 1.0f);
    float w = fminf(accW[px], 1.0f);

    float thr = hash01(px, frameCounter); // 0..1

    float r255 = r * 255.0f;
    uint8_t r8 = (uint8_t)floorf(r255 + ((r255 - floorf(r255)) > thr ? 1.0f : 0.0f));
    float g255 = g * 255.0f;
    uint8_t g8 = (uint8_t)floorf(g255 + ((g255 - floorf(g255)) > thr ? 1.0f : 0.0f));
    float b255 = b * 255.0f;
    uint8_t b8 = (uint8_t)floorf(b255 + ((b255 - floorf(b255)) > thr ? 1.0f : 0.0f));
    float w255 = w * 255.0f;
    uint8_t w8 = (uint8_t)floorf(w255 + ((w255 - floorf(w255)) > thr ? 1.0f : 0.0f));

    r8 = leds.gamma8(r8);
    g8 = leds.gamma8(g8);
    b8 = leds.gamma8(b8);
    w8 = leds.gamma8(w8);

    leds.setPixelColor(px, packColor(r8, g8, b8, w8));
  }
}

// ======== ARDUINO ========
void setup()
{
  randomSeed(analogRead(28));
#ifdef PIN_NEOPIXEL_POWER
  pinMode(PIN_NEOPIXEL_POWER, OUTPUT);
  digitalWrite(PIN_NEOPIXEL_POWER, LOW);
#endif

  if (!leds.begin())
  {
    pinMode(LED_BUILTIN, OUTPUT);
    for (;;)
      digitalWrite(LED_BUILTIN, (millis() / 200) & 1);
  }
  leds.setBrightness(255);
  leds.clear();
  leds.show();

  initTwinkles();
  pulses.clear();
  spawnPulse(); // start with one pulse
  lastFrameMs = millis();
}

void loop()
{
  uint32_t now = millis();
  float dt = (now - lastFrameMs) / 1000.0f;
  if (dt < 0.001f)
    dt = 0.001f;
  if (dt > 0.050f)
    dt = 0.050f;

  renderFrame(dt);
  leds.show();

  frameCounter++; // advance dithering phase
  lastFrameMs = now;
  delay(16); // ~60 FPS
}
