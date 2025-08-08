# Configure board for Arduino

1. Arduino > Preferences
2. Add https://github.com/earlephilhower/arduino-pico/releases/download/global/package_rp2040_index.json
3. Boards Manager -> Raspberry Pi Pico/RP2040
4. Boards -> Adafruit Feather RP2040 Scorpio
5. Sketch -> Include Library -> Manage Libraries ->
  - Adafruit_NeoPXL8
  - Adafruit_NeoPixel
  - Adafruit_ZeroDMA

# First time setup

```bash
brew install arm-none-eabi-gcc cmake ninja
which arm-none-eabi-gcc
arm-none-eabi-gcc --version
```

```bash
# download https://developer.arm.com/-/media/Files/downloads/gnu/13.3.rel1/binrel/arm-gnu-toolchain-13.3.rel1-darwin-arm64-arm-none-eabi.tar.xz
# download

# download https://developer.arm.com/-/media/Files/downloads/gnu/14.3.rel1/binrel/arm-gnu-toolchain-14.3.rel1-darwin-arm64-arm-none-eabi.tar.xz

mkdir -p $HOME/opt && tar -C $HOME/opt -xf ~/Downloads/arm-gnu-toolchain-14.3.rel1-darwin-arm64-arm-none-eabi.tar.xz
export PATH="$HOME/opt/arm-gnu-toolchain-14.3.rel1-darwin-arm64-arm-none-eabi/bin:$PATH"
```



# Build

```bash
export PICO_SDK_PATH="$HOME/code/pico-sdk"\
export PATH="$(brew --prefix)/bin:$PATH"

cp $PICO_SDK_PATH/external/pico_sdk_import.cmake .
mkdir -p build && cd build
cmake .. -G Ninja -DPICO_SDK_PATH=$PICO_SDK_PATH
ninja
```
