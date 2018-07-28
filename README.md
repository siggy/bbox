# BBox

Beatboxer in Go

## Raspberry PI Setup

### OS

1. Download Raspian Lite: https://downloads.raspberrypi.org/raspbian_lite_latest
2. Flash `2017-07-05-raspbian-jessie-lite.zip` using Etcher
3. Remove/reinsert flash drive
4. Add `ssh` file:
```bash
touch /Volumes/boot/ssh
diskutil umount /Volumes/boot
```

### First Boot

```bash
ssh pi@raspberrypi.local
# password: raspberry

# change default password
passwd

# set quiet boot
sudo sed -i '${s/$/ quiet loglevel=1/}' /boot/cmdline.txt

# install packages
sudo apt-get update
sudo apt-get install -y git tmux vim dnsmasq hostapd

# set up wifi (note leading space to avoid bash history)
 echo $'\nnetwork={\n    ssid="<WIFI_SSID>"\n    psk="<WIFI_PASSWORD>"\n}' | sudo tee --append /etc/wpa_supplicant/wpa_supplicant.conf

# set static IP address
echo $'\n# set static ip\n\ninterface eth0\nstatic ip_address=192.168.1.141/24\nstatic routers=192.168.1.1\nstatic domain_name_servers=192.168.1.1\n\ninterface wlan0\nstatic ip_address=192.168.1.142/24\nstatic routers=192.168.1.1\nstatic domain_name_servers=192.168.1.1' | sudo tee --append /etc/dhcpcd.conf

# to make as a wifi access point at 192.168.4.1
echo $'\ninterface wlan0\nnohook wpa_supplicant' | sudo tee --append /etc/dhcpcd.conf
#echo $'\ndenyinterfaces wlan0' | sudo tee --append /etc/dhcpcd.conf
echo $'\n# set static ip\n\ninterface eth0\nstatic ip_address=192.168.4.1/24\nstatic routers=192.168.1.1\nstatic domain_name_servers=192.168.1.1\n\ninterface wlan0\nstatic ip_address=192.168.1.142/24\nstatic routers=192.168.1.1\nstatic domain_name_servers=192.168.1.1\nnohook wpa_supplicant' | sudo tee --append /etc/dhcpcd.conf
echo $'\ninterface=wlan0\ndhcp-range=192.168.4.2,192.168.4.20,255.255.255.0,24h' | sudo tee --append /etc/dnsmasq.conf
sudo tee /etc/hostapd/hostapd.conf > /dev/null <<'EOF'
interface=wlan0
driver=nl80211
ssid=sigpi
hw_mode=g
channel=7
wmm_enabled=0
macaddr_acl=0
auth_algs=1
ignore_broadcast_ssid=0
wpa=2
wpa_passphrase=showmethepi
wpa_key_mgmt=WPA-PSK
wpa_pairwise=TKIP
rsn_pairwise=CCMP
EOF
echo $'\nDAEMON_CONF="/etc/hostapd/hostapd.conf"' | sudo tee --append /etc/default/hostapd
echo $'\nnet.ipv4.ip_forward=1' | sudo tee --append /etc/sysctl.conf
sudo iptables -t nat -A  POSTROUTING -o eth0 -j MASQUERADE
sudo sh -c "iptables-save > /etc/iptables.ipv4.nat"
# add to end of /etc/rc.local:
# iptables-restore < /etc/iptables.ipv4.nat
sudo systemctl start hostapd
sudo systemctl start dnsmasq

# reboot to connect over wifi
sudo shutdown -r now

# configure git
git config --global push.default simple
git config --global core.editor "vim"
git config --global user.email "you@example.com"
git config --global user.name "Your Name"

# disable services
sudo systemctl disable hciuart
sudo systemctl disable bluetooth
sudo systemctl disable plymouth

# remove unnecessary packages
sudo apt-get -y purge libx11-6 libgtk-3-common xkb-data lxde-icon-theme raspberrypi-artwork penguinspuzzle ntp plymouth*
sudo apt-get -y autoremove

sudo raspi-config nonint do_boot_behaviour B2 0
sudo raspi-config nonint do_boot_wait 1
sudo raspi-config nonint do_serial 1
```

## Code

```bash
wget https://dl.google.com/go/go1.10.3.linux-armv6l.tar.gz -O /tmp/go1.10.3.linux-armv6l.tar.gz
sudo tar -xzf /tmp/go1.10.3.linux-armv6l.tar.gz -C /usr/local

mkdir -p ~/code/go/src/github.com/siggy
git clone https://github.com/siggy/bbox.git ~/code/go/src/github.com/siggy/bbox
```

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
go build -o beatboxer cmd/bbox.go && \
    go build cmd/amplitude.go && \
    go build cmd/aud.go &&       \
    go build cmd/aux.go &&       \
    go build cmd/clear.go &&     \
    go build cmd/crane.go &&     \
    go build cmd/crawler.go &&   \
    go build cmd/fish.go &&      \
    go build cmd/keys.go &&      \
    go build cmd/leds.go &&      \
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

## Editing SD card

Launch Ubuntu in VirtualBox

```bash
sudo mount /dev/sdb7 ~/usb
sudo umount /dev/sdb7
```

## Docs

```bash
jekyll serve -s docs
open http://127.0.0.1:4000/bbox
```

## Credits

- [wavs](wavs) courtesy of (http://99sounds.org/drum-samples/)
- [rpi_ws281x](rpi_ws281x) courtesy of (https://github.com/jgarff/rpi_ws281x)
