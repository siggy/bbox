# TODO

- turn up volume output from bbox pi

# BBox

Beatboxer in Go

- [TODO](#todo)
- [BBox](#bbox)
  - [Raspberry PI Setup](#raspberry-pi-setup)
    - [OS](#os)
    - [First Boot](#first-boot)
    - [Faster boot](#faster-boot)
    - [Code](#code)
    - [USB Audio](#usb-audio)
      - [bbox-style sound card](#bbox-style-sound-card)
      - [baux-style sound card](#baux-style-sound-card)
      - [Test](#test)
    - [Audio support for Go](#audio-support-for-go)
  - [Run](#run)
  - [Build](#build)
  - [Auto boot with keyboard attach](#auto-boot-with-keyboard-attach)
    - [Auto boot baux](#auto-boot-baux)
  - [Connectivity](#connectivity)
    - [Configure to connect over ethernet](#configure-to-connect-over-ethernet)
    - [Configure the Pi to connect as Wifi AP](#configure-the-pi-to-connect-as-wifi-ap)
      - [Switch the Pi back to connecting to the internet with wifi](#switch-the-pi-back-to-connecting-to-the-internet-with-wifi)
  - [Docs](#docs)
  - [Credits](#credits)


## Raspberry PI Setup

### OS

1. Install Raspberry Pi Imager: https://www.raspberrypi.com/software/
2. Choose `Raspberry Pi 5`, `Raspberry Pi OS Lite (64-bit)` and
   `USB Mass Storage` in the Imager
3. Select:
  - `Set hostname` to `raspberrypi`
  - `Set username and password`: `sig`
  - `Set locale` to `en_US.UTF-8`
  - `Configure wireless LAN`
  - `Enable SSH`: `Allow public-key authentication only`

### First Boot

```bash
HOSTNAME=raspberrypi
HOSTNAME=raspberrypi5-2
HOSTNAME=raspberrypi5-3
HOSTNAME=raspberrypi5-4
HOSTNAME=raspberrypi5-5
HOSTNAME=raspberrypi5-6

ssh sig@raspberrypi.local
ssh sig@$HOSTNAME.local
ssh sig@raspberrypi5-5.local

# install packages
sudo apt-get update
sudo sed -i 's/^# *\(en_US.UTF-8 UTF-8\)/\1/' /etc/locale.gen
sudo locale-gen
sudo dpkg-reconfigure locales
sudo apt-get install -y git tmux vim locales

# GITHUB_TOKEN=

git clone https://github.com/siggy/dotfiles.git ~/code/dotfiles
cp ~/code/dotfiles/.local.bash.pi ~/.local.bash
cp ~/code/dotfiles/.curlrc        ~/
cp ~/code/dotfiles/.gitconfig     ~/
cp ~/code/dotfiles/.tmux.conf     ~/
cp ~/code/dotfiles/.vimrc         ~/
cp ~/code/dotfiles/.wgetrc        ~/

sed -i '/^[[:space:]]*helper = osxkeychain/ s/^/#/' ~/.gitconfig

mkdir -p ~/.vim/backups
mkdir -p ~/.vim/swaps
cp -a ~/code/dotfiles/.vim/colors ~/.vim

echo "[[ -s ${HOME}/.local.bash ]] && source ${HOME}/.local.bash" >> ~/.profile
```

### Faster boot

```bash
sudo systemctl disable ModemManager.service
sudo systemctl disable NetworkManager-wait-online.service
sudo systemctl disable bluetooth.service
sudo systemctl disable dphys-swapfile.service
sudo systemctl disable fake-hwclock.service
sudo systemctl disable systemd-binfmt.service
sudo systemctl mask sys-kernel-debug.mount sys-kernel-tracing.mount
sudo systemctl mask rpi-eeprom-update
sudo systemctl disable e2scrub_reap.service
sudo dphys-swapfile swapoff
sudo apt-get -y purge modemmanager bluez triggerhappy
sudo apt-get -y autoremove --purge

sudo grep -q 'rootdelay=' /boot/firmware/cmdline.txt \
  || sudo sed -i 's/$/ rootdelay=2 modules-load=dwc2/' /boot/firmware/cmdline.txt

echo "PollIntervalMinSec=600" | sudo tee -a /etc/systemd/timesyncd.conf
```

### Code

```bash
curl -L -o /tmp/go1.24.4.linux-arm64.tar.gz https://go.dev/dl/go1.24.4.linux-arm64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go1.24.4.linux-arm64.tar.gz
export PATH=$PATH:/usr/local/go/bin

mkdir -p ~/code/
git clone https://github.com/siggy/bbox.git ~/code/bbox
```

### USB Audio

#### bbox-style sound card

For MCSPER USB to 3.5mm Audio Jack Adapter.

Plug in USB audio device and run:
```bash
sudo tee /etc/asound.conf > /dev/null <<'EOF'
pcm.!default {
  type plug
  slave.pcm "dmix:Audio,0"
}

ctl.!default {
  type hw
  card "Audio"
}
EOF

amixer
amixer -c 0 scontrols
amixer -c 0 set PCM 95%
```

#### baux-style sound card

For Plugable USB Audio Adapter with 3.5mm Speaker-Headphone and Microphone Jack.

Plug in USB audio device and run:
```bash
sudo tee /etc/asound.conf > /dev/null <<'EOF'
pcm.!default {
  type plug
  slave.pcm "dmix:Device,0"
}

ctl.!default {
  type hw
  card "Device"
}
EOF
```

#### Test

```bash
aplay /usr/share/sounds/alsa/Front_Center.wav
```

### Audio support for Go

```bash
sudo apt-get install -y libasound2-dev
```

## Run

```bash
go run cmd/bbox2/main.go --fake-leds
```

## Build

```bash
go build -o /home/sig/bin/bbox cmd/bbox2/main.go
```

## Auto boot with keyboard attach

```bash
cat <<'EOF' >> ~/.profile

if [ "$(tty)" = "/dev/tty1" ]; then
  tmux attach -t bbox || tmux new-session -s bbox "bash -c '/home/sig/bin/bbox; exec bash'"
fi
EOF
```

```bash
sudo mkdir -p /etc/systemd/system/getty@tty1.service.d
cat <<EOF | sudo tee /etc/systemd/system/getty@tty1.service.d/override.conf
[Service]
ExecStart=
ExecStart=-/sbin/agetty --autologin sig --noclear %I \$TERM
EOF

sudo systemctl daemon-reload
sudo reboot
```

### Auto boot baux

```bash
cat <<'EOF' >> ~/.profile

export MINIAUDIO_ALSA_NO_MMAP=1
if [ "$(tty)" = "/dev/tty1" ]; then
  tmux attach -t bbox || tmux new-session -s bbox "bash -c '/home/sig/bin/baux; exec bash'"
fi
EOF
```

## Connectivity

### Configure to connect over ethernet

```bash
IP=192.168.2.2
IP=192.168.2.3
IP=192.168.2.4
IP=192.168.2.5
IP=192.168.2.6
sudo nmcli con add type ethernet ifname eth0 con-name eth0-static ipv4.addresses $IP/24 ipv4.method manual
sudo nmcli con up eth0-static
```

```bash
HOSTNAME=raspberrypi
HOSTNAME=raspberrypi5-2
HOSTNAME=raspberrypi5-3
HOSTNAME=raspberrypi5-4
HOSTNAME=raspberrypi5-5
HOSTNAME=raspberrypi5-6
IP=192.168.2.2
IP=192.168.2.2
IP=192.168.2.3
IP=192.168.2.4
IP=192.168.2.5
IP=192.168.2.6
alias pi="ssh sig@$HOSTNAME.local"
alias pieth="ssh sig@$IP"
```

### Configure the Pi to connect as Wifi AP

```bash
SSID=sigpi
SSID=sigpi5-2
SSID=sigpi5-3
SSID=sigpi5-4
SSID=sigpi5-5
SSID=sigpi5-6
rfkill list
sudo rfkill unblock wifi
rfkill list
sudo nmcli radio wifi on
sudo nmcli dev wifi hotspot ifname wlan0 ssid $SSID password showmethepi
nmcli dev wifi show-password
sudo nmcli connection modify Hotspot autoconnect yes
```

From laptop:
```bash
HOSTNAME=raspberrypi5-4
ssh sig@$HOSTNAME.local
```


#### Switch the Pi back to connecting to the internet with wifi

```bash
nmcli connection show
sudo nmcli connection down Hotspot
```

## Docs

```bash
jekyll serve -s docs
open http://127.0.0.1:4000/bbox
```

## Credits

- [wavs](wavs) courtesy of (http://99sounds.org/drum-samples/)
- Keyboard courtesy of (https://github.com/kodachi614/macropaw)
