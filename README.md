# TODO

- confirm drum timing
- handle multiple keyboards
- fix baux deadlock
- guarantee scorpio at /dev/ttyACM0
- guarantee audio at card:
  defaults.pcm.card 1
  defaults.ctl.card 1

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
    - [Audio support for Go](#audio-support-for-go)
  - [Run](#run)
  - [Build](#build)
  - [Auto boot with keyboard attach](#auto-boot-with-keyboard-attach)
  - [Connectivity](#connectivity)
    - [Configure to connect over ethernet](#configure-to-connect-over-ethernet)
    - [Configure the Pi to connect as Wifi AP](#configure-the-pi-to-connect-as-wifi-ap)
      - [Switch the Pi back to connecting to the internet with wifi](#switch-the-pi-back-to-connecting-to-the-internet-with-wifi)
- [OLD](#old)
    - [portaudio](#portaudio)
    - [rpi\_ws281x](#rpi_ws281x)
  - [Env / bootup](#env--bootup)
  - [Build](#build-1)
  - [Run](#run-1)
  - [Stop bbox process](#stop-bbox-process)
  - [Check for voltage drop](#check-for-voltage-drop)
  - [Editing SD card](#editing-sd-card)
  - [Wifi access point](#wifi-access-point)
    - [To re-enable internet wifi](#to-re-enable-internet-wifi)
    - [Mounting / syncing pi](#mounting--syncing-pi)
    - [Set restart crontab](#set-restart-crontab)
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

Plug in USB audio device and run:
```bash
# note the card number of the USB audio device
aplay -l

sudo tee /etc/asound.conf > /dev/null <<'EOF'
defaults.pcm.card 1
defaults.ctl.card 1
EOF

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
HOSTNAME=raspberrypi5-3
HOSTNAME=raspberrypi5-4
HOSTNAME=raspberrypi5-5
HOSTNAME=raspberrypi5-6
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












# OLD


### portaudio

```bash
# OSX
brew install portaudio

# Raspbian
sudo apt-get install -y libasound-dev

wget http://portaudio.com/archives/pa_stable_v190600_20161030.tgz -O /tmp/pa_stable_v190600_20161030.tgz
cd /tmp
tar -xzf pa_stable_v190600_20161030.tgz
cd portaudio
./configure
make
sudo make install
sudo ldconfig
```

### rpi_ws281x

Beatboxer depends on a fork of (https://github.com/jgarff/rpi_ws281x). See that
repo for complete instructions.

```bash
cd ~/code/go/src/github.com/siggy/bbox
sudo cp rpi_ws281x/rpihw.h  /usr/local/include/
sudo cp rpi_ws281x/ws2811.h /usr/local/include/
sudo cp rpi_ws281x/pwm.h    /usr/local/include/

sudo cp rpi_ws281x/libws2811.a /usr/local/lib/

# osx
export CGO_CFLAGS="$CGO_CFLAGS -I/usr/local/include"
export CGO_LDFLAGS="$CGO_LDFLAGS -L/usr/local/lib"
```

## Env / bootup

```bash
# set bootup and shell env
cd ~/code/go/src/github.com/siggy/bbox
cp rpi/.local.bash ~/
source ~/.local.bash

cp rpi/bboxgo.sh ~/
sudo cp rpi/bbox.service /etc/systemd/system/bbox.service
sudo systemctl enable bbox

echo "[[ -s ${HOME}/.local.bash ]] && source ${HOME}/.local.bash" >> ~/.bashrc

# audio setup

# external sound card
sudo cp rpi/asound.conf /etc/

# *output of raspi-config after forcing audio to hdmi*
numid=3,iface=MIXER,name='Mic Playback Switch'
  ; type=BOOLEAN,access=rw------,values=1
  : values=on
# *also this might work*
amixer cset numid=3 2
# OR:
sudo raspi-config nonint do_audio 2

echo "blacklist snd_bcm2835" | sudo tee --append /etc/modprobe.d/snd-blacklist.conf

echo "hdmi_force_hotplug=1" | sudo tee --append /boot/config.txt
echo "hdmi_force_edid_audio=1" | sudo tee --append /boot/config.txt

# make usb audio card #0
sudo vi /lib/modprobe.d/aliases.conf
#options snd-usb-audio index=-2

# reboot

aplay -l
# ... should match the contents of asound.conf, and also:
sudo vi /usr/share/alsa/alsa.conf
defaults.ctl.card 0
defaults.pcm.card 0
```

## Build

```bash
go build cmd/beatboxer_noleds.go && \
  go build cmd/beatboxer_leds.go && \
  go build cmd/baux.go &&      \
  go build cmd/clear.go &&     \
  go build cmd/fishweb.go &&   \
  go build cmd/human.go &&      \
  go build cmd/leds.go &&      \

  go build cmd/amplitude.go && \
  go build cmd/aud.go &&       \
  go build cmd/crane.go &&     \
  go build cmd/crawler.go &&   \
  go build cmd/fish.go &&      \
  go build cmd/keys.go &&      \
  go build cmd/noleds.go &&    \
  go build cmd/record.go
```

## Run

All programs that use LEDs must be run with `sudo`.

```bash
sudo ./beatboxer # main program
sudo ./leds # led testing
sudo ./clear # clear LEDs
./noleds # beatboxer without LEDs (for testing without pi)
./aud # audio testing
./keys # keyboard test
```

## Stop bbox process

```bash
# the systemd way
sudo systemctl stop bbox

# send SIGINT to turn off LEDs
sudo kill -2 <PID>
```

## Check for voltage drop

```bash
vcgencmd get_throttled
```

## Editing SD card

Launch Ubuntu in VirtualBox

```bash
sudo mount /dev/sdb7 ~/usb
sudo umount /dev/sdb7
```

## Wifi access point

Based on:
https://frillip.com/using-your-raspberry-pi-3-as-a-wifi-access-point-with-hostapd/

```bash
sudo tee --append /etc/dhcpcd.conf > /dev/null <<'EOF'

# this must go above any `interface` line
denyinterfaces wlan0

# this must go below `interface wlan0`
nohook wpa_supplicant
EOF

sudo tee --append /etc/network/interfaces > /dev/null <<'EOF'

allow-hotplug wlan0
iface wlan0 inet static
    address 192.168.4.1
    netmask 255.255.255.0
    network 192.168.4.0
    broadcast 192.168.1.255
#    wpa-conf /etc/wpa_supplicant/wpa_supplicant.conf
EOF

sudo tee /etc/hostapd/hostapd.conf > /dev/null <<'EOF'
interface=wlan0
driver=nl80211
ssid=sigpi
hw_mode=g
channel=6
ieee80211n=1
wmm_enabled=0
ht_capab=[HT40][SHORT-GI-20][DSSS_CCK-40]
macaddr_acl=0
auth_algs=1
ignore_broadcast_ssid=0
wpa=2
wpa_key_mgmt=WPA-PSK
wpa_passphrase=showmethepi
# wpa_pairwise=TKIP
rsn_pairwise=CCMP
EOF

sudo tee /etc/default/hostapd > /dev/null <<'EOF'
DAEMON_CONF="/etc/hostapd/hostapd.conf"
EOF

sudo tee --append /etc/dnsmasq.conf > /dev/null <<'EOF'

interface=wlan0
listen-address=192.168.4.1
bind-interfaces
domain-needed
dhcp-range=192.168.4.2,192.168.4.100,255.255.255.0,24h
EOF

sudo tee --append /etc/sysctl.conf > /dev/null <<'EOF'

net.ipv4.ip_forward=1
EOF

sudo service dhcpcd restart
sudo systemctl start hostapd
sudo systemctl start dnsmasq

# reboot to connect over wifi
sudo shutdown -r now
```

### To re-enable internet wifi

Comment out from `/etc/dhcpcd.conf`:
```
# denyinterfaces wlan0
# nohook wpa_supplicant
```

Re-enable in `/etc/network/interfaces`:
```
allow-hotplug wlan0
iface wlan0 inet manual
    wpa-conf /etc/wpa_supplicant/wpa_supplicant.conf
```

sudo service dhcpcd restart
sudo ifdown wlan0; sudo ifup wlan0
sudo systemctl stop hostapd
sudo systemctl stop dnsmasq

### Mounting / syncing pi

```bash
# mount pi volume locally
alias pifs='umount /Volumes/pi; sudo rmdir /Volumes/pi; sudo mkdir /Volumes/pi; sudo chown sig:staff /Volumes/pi && sshfs pi@raspberrypi.local:/ /Volumes/pi -f'

# rsync local repo to pi
rsync -vr ~/code/go/src/github.com/siggy/bbox/.git /Volumes/pi/home/pi/code/go/src/github.com/siggy/bbox

# remove volume mount
alias pifsrm='umount /Volumes/pi; sudo rmdir /Volumes/pi; sudo mkdir /Volumes/pi; sudo chown sig:staff /Volumes/pi'
```

### Set restart crontab

```bash
sudo crontab -e
```

```
*/10 * *   *   *     sudo /sbin/shutdown -r now
```

```
sudo crontab -l
```

## Docs

```bash
jekyll serve -s docs
open http://127.0.0.1:4000/bbox
```

## Credits

- [wavs](wavs) courtesy of (http://99sounds.org/drum-samples/)
- [rpi_ws281x](rpi_ws281x) courtesy of (https://github.com/jgarff/rpi_ws281x)
