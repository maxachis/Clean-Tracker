const BEEP_HZ = 880;
const BEEP_SECONDS = 0.18;
const BEEP_GAP_SECONDS = 0.1;
const BEEP_COUNT = 2;
const BEEP_GAIN = 0.25;

let context: AudioContext | null = null;

/** Two short Web Audio beeps; silently does nothing where audio is unavailable. */
export function playMissBeep(): void {
  try {
    const Ctor = window.AudioContext;
    if (!Ctor) return;
    context ??= new Ctor();
    const ctx = context;
    for (let i = 0; i < BEEP_COUNT; i++) {
      scheduleBeep(ctx, ctx.currentTime + i * (BEEP_SECONDS + BEEP_GAP_SECONDS));
    }
  } catch {
    // Audio is a nicety; never break the session over it.
  }
}

function scheduleBeep(ctx: AudioContext, startAt: number): void {
  const osc = ctx.createOscillator();
  const gain = ctx.createGain();
  osc.type = "square";
  osc.frequency.value = BEEP_HZ;
  gain.gain.value = BEEP_GAIN;
  osc.connect(gain).connect(ctx.destination);
  osc.start(startAt);
  osc.stop(startAt + BEEP_SECONDS);
}
