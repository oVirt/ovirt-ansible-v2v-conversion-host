lang en_US
keyboard us
timezone America/New_York --isUtc

rootpw --lock

#platform x86, AMD64, or Intel EM64T
reboot
text
cdrom
bootloader --location=mbr --append="rhgb quiet crashkernel=auto"
zerombr
clearpart --all --initlabel
autopart
auth --passalgo=sha512 --useshadow
selinux --enforcing
firewall --enabled --ssh
firstboot --disable
network --bootproto=dhcp --onboot=on

repo --name=base --baseurl=http://mirror.centos.org/centos/7/os/x86_64/
repo --name=epel --baseurl=http://dl.fedoraproject.org/pub/epel/7/x86_64/

%packages
ansible
cloud-init
-NetworkManager

# tasks/install.yml
nbdkit
nbdkit-plugin-python2
virt-v2v

# tasks/install-openstack.yml
python-openstackclient
python-six
%end

%post --log=/root/ks-post-v2v.log --erroronfail
echo "Installing oVirt repo.."
yum install -y http://resources.ovirt.org/pub/yum-repo/ovirt-release42.rpm

echo "Installing ovirt-ansible-v2v-conversion-host package.."
yum install -y ovirt-ansible-v2v-conversion-host

echo "Downloading virtio-win ISO"
mkdir -p /usr/share/virtio-win
curl -vL -o /usr/share/virtio-win/virtio-win.iso 'https://fedorapeople.org/groups/virt/virtio-win/direct-downloads/stable-virtio/virtio-win.iso'

echo -n "Network fixes"
# initscripts don't like this file to be missing.
cat > /etc/sysconfig/network << EOF
NETWORKING=yes
NOZEROCONF=yes
EOF

# For cloud images, 'eth0' _is_ the predictable device name, since
# we don't want to be tied to specific virtual (!) hardware
rm -f /etc/udev/rules.d/70*
ln -s /dev/null /etc/udev/rules.d/80-net-name-slot.rules

# simple eth0 config, again not hard-coded to the build hardware
cat > /etc/sysconfig/network-scripts/ifcfg-eth0 << EOF
DEVICE="eth0"
BOOTPROTO="dhcp"
BOOTPROTOv6="dhcp"
ONBOOT="yes"
TYPE="Ethernet"
USERCTL="yes"
PEERDNS="yes"
IPV6INIT="yes"
PERSISTENT_DHCLIENT="1"
EOF
%end
