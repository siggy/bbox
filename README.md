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
```

### First Boot

```bash
ssh pi@raspberrypi.local
# password: raspberry

# change default password
passwd

# set quiet boot
sudo sed -i '${s/$/ quiet loglevel=1/}' /boot/cmdline.txt

# set up wifi (note leading space to avoid bash history)
 echo $'\nnetwork={\n    ssid="<WIFI_SSID>"\n    psk="<WIFI_PASSWORD>"\n}' | sudo tee --append /etc/wpa_supplicant/wpa_supplicant.conf

# set static IP address
echo $'\n# set static ip\n\ninterface eth0\nstatic ip_address=192.168.1.141/24\nstatic routers=192.168.1.1\nstatic domain_name_servers=192.168.1.1\n\ninterface wlan0\nstatic ip_address=192.168.1.142/24\nstatic routers=192.168.1.1\nstatic domain_name_servers=192.168.1.1' | sudo tee --append /etc/dhcpcd.conf

# reboot to connect over wifi
sudo shutdown -r now

# install packages
sudo apt-get update
sudo apt-get install -y git tmux vim

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
wget https://storage.googleapis.com/golang/go1.8.3.linux-armv6l.tar.gz -O /tmp/go1.8.3.linux-armv6l.tar.gz
sudo tar -xzf /tmp/go1.8.3.linux-armv6l.tar.gz -C /usr/local

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
sudo cp rpi_ws281x/rpihw.h  /usr/local/include/
sudo cp rpi_ws281x/ws2811.h /usr/local/include/
sudo cp rpi_ws281x/pwm.h    /usr/local/include/

sudo cp rpi_ws281x/libws2811.a /usr/local/lib/

# osx
export CGO_CFLAGS="$CGO_CFLAGS -I/usr/local/include"
export CGO_LDFLAGS="$CGO_LDFLAGS -L/usr/local/lib"
```

## Build

```bash
go build -o beatboxer cmd/bbox.go && go build cmd/amplitude.go && go build cmd/aud.go && go build cmd/clear.go && go build cmd/crane.go && go build cmd/crawler.go && go build cmd/fish.go && go build cmd/keys.go && go build cmd/leds.go && go build cmd/noleds.go && go build cmd/record.go
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

## Env / bootup

```bash
# set bootup and shell env
cp ~/code/go/src/github.com/siggy/bbox/rpi/.local.bash ~/
source ~/.local.bash
cd ~/code/go/src/github.com/siggy/bbox/
go build ~/code/go/src/github.com/siggy/bbox/cmd/clear.go
go build ~/code/go/src/github.com/siggy/bbox/cmd/fish.go

cp ~/code/go/src/github.com/siggy/bbox/rpi/bboxgo.sh ~/
sudo cp ~/code/go/src/github.com/siggy/bbox/rpi/bbox.service /etc/systemd/system/bbox.service
sudo systemctl enable bbox

echo "[[ -s ${HOME}/.local.bash ]] && source ${HOME}/.local.bash" >> ~/.bashrc

# audio setup

# external sound card
sudo cp ~/code/go/src/github.com/siggy/bbox/rpi/asound.conf /etc/

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

## Credits

- [wavs](wavs) courtesy of (http://99sounds.org/drum-samples/)
- [rpi_ws281x](rpi_ws281x) courtesy of (https://github.com/jgarff/rpi_ws281x)
