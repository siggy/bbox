# Setup for Raspberry PI

## OS Setup

1. Download Raspian Lite: https://downloads.raspberrypi.org/raspbian_lite_latest
2. Flash `2017-07-05-raspbian-jessie-lite.zip` using Etcher
3. Add `ssh` file:
```bash
touch /Volumes/boot/ssh
```

### In Raspbian GUI

## First Boot

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

## Env / bootup

```bash
# external sound card
sudo cp ~/code/go/src/github.com/siggy/bbox/rpi/asound.conf /etc/

# set bootup and shell env
cp ~/code/go/src/github.com/siggy/bbox/rpi/.local.bash ~/
cp ~/code/go/src/github.com/siggy/bbox/rpi/bboxgo.sh ~/
sudo cp ~/code/go/src/github.com/siggy/bbox/rpi/bbox.service /etc/systemd/system/bbox.service
sudo systemctl enable bbox

echo "[[ -s ${HOME}/.local.bash ]] && source ${HOME}/.local.bash" >> ~/.bashrc
```

# *output of raspi-config after forcing audio to hdmi*
numid=3,iface=MIXER,name='Mic Playback Switch'
  ; type=BOOLEAN,access=rw------,values=1
  : values=on
# *also this might work*
amixer cset numid=3 2

## Editing SD card

Launch Ubuntu in VirtualBox

```bash
sudo mount /dev/sdb7 ~/usb
sudo umount /dev/sdb7
```
