import sys
import numpy as np
import onnxruntime
import struct
import logging
import torch
import librosa

# --- Configuration ---
# These parameters MUST match the model's training
N_FFT = 512
HOP_LENGTH = 256
WIN_LENGTH = 512
SAMPLE_RATE = 16000 # The model expects 16kHz audio

# Set up logging to stderr to not interfere with stdout data pipe
logging.basicConfig(level=logging.INFO, stream=sys.stderr, format='%(levelname)s: %(message)s')

def process_audio():
    """
    Initializes the ONNX model and enters a loop to process raw audio frames
    received from stdin and writes the enhanced raw audio frames to stdout.
    """
    try:
        logging.info(f"Initializing ONNX session for gtcrn_simple.onnx...")
        session = onnxruntime.InferenceSession('gtcrn_simple.onnx', None, providers=['CPUExecutionProvider'])
        logging.info("ONNX session initialized successfully.")
    except Exception as e:
        logging.error(f"Failed to load ONNX model 'gtcrn_simple.onnx': {e}")
        return

    # Initialize model state caches
    conv_cache = np.zeros([2, 1, 16, 16, 33], dtype="float32")
    tra_cache = np.zeros([2, 3, 1, 1, 16], dtype="float32")
    inter_cache = np.zeros([2, 1, 33, 16], dtype="float32")

    logging.info("Python service is ready and waiting for raw audio data...")
    hann_window = torch.hann_window(WIN_LENGTH).pow(0.5)

    while True:
        try:
            # 1. Read raw audio bytes from Go
            len_bytes = sys.stdin.buffer.read(4)
            if not len_bytes:
                logging.info("Stdin closed, exiting.")
                break
            
            frame_len = struct.unpack('<I', len_bytes)[0]
            frame_bytes = sys.stdin.buffer.read(frame_len)
            if len(frame_bytes) < frame_len:
                logging.warning("Incomplete frame received, exiting.")
                break

            # 2. Convert bytes (int16) to float32 tensor
            raw_samples = np.frombuffer(frame_bytes, dtype=np.int16).astype(np.float32) / 32768.0
            
            
            # Resample if necessary. Your WAVs are 44.1kHz, model needs 16kHz
            if len(raw_samples) > 0:
                raw_samples = librosa.resample(raw_samples, orig_sr=44100, target_sr=SAMPLE_RATE)
            
            x = torch.from_numpy(raw_samples)

            # 3. Perform STFT
            spectrogram = torch.stft(x, N_FFT, HOP_LENGTH, WIN_LENGTH, hann_window, return_complex=False)[None]
            inputs = spectrogram.numpy()

            # 4. Run Inference
            outputs = []
            for i in range(inputs.shape[-2]):
                out_i, conv_cache, tra_cache, inter_cache = session.run(
                    [], 
                    {
                        'mix': inputs[..., i:i + 1, :],
                        'conv_cache': conv_cache,
                        'tra_cache': tra_cache,
                        'inter_cache': inter_cache
                    }
                )
                outputs.append(out_i)

            # 5. Perform inverse STFT
            enhanced_spec = np.concatenate(outputs, axis=2)
            enhanced_audio_complex = enhanced_spec[..., 0] + 1j * enhanced_spec[..., 1]
            enhanced_audio = librosa.istft(enhanced_audio_complex, hop_length=HOP_LENGTH, win_length=WIN_LENGTH, window='hann')

            # Resample back to 44.1kHz to match the output device
            # enhanced_audio = librosa.resample(enhanced_audio.squeeze(), orig_sr=SAMPLE_RATE, target_sr=44100)

            # 6. Convert float32 samples back to int16 bytes
            enhanced_audio_int16 = (enhanced_audio * 32767).astype(np.int16)
            output_bytes = enhanced_audio_int16.tobytes()

            # 7. Write processed bytes back to Go
            sys.stdout.buffer.write(struct.pack('<I', len(output_bytes)))
            sys.stdout.buffer.write(output_bytes)
            sys.stdout.buffer.flush()

        except Exception as e:
            logging.error(f"An error occurred during processing: {e}", exc_info=True)
            break

if __name__ == "__main__":
    process_audio()