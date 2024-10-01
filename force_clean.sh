#!/usr/bin/env bash

ETCD=etcd
KUBE_APISERVER=kube-apiserver
KUBE_CONTROLLER_MANAGER=kube-controller-manager
KUBE_SCHEDULER=kube-scheduler
KUBELET=kubelet
KUBEPROXY=kube-proxy
KUBERNETES_CNI=kubernetes-cni
KUBEADM=kubeadm
CONTAINERD=containerd
KUBECTL=kubectl
RUNC=runc
CRICTL=crictl

OS_DISTRO=$(grep -w ID /etc/os-release | cut -d= -f2 | tr -d '"')
PACKAGE_MANAGER=$([ -e '/usr/bin/yum' ] && echo "yum" || echo "zypper")

for i in ${ETCD} ${KUBE_APISERVER} ${KUBE_CONTROLLER_MANAGER} ${KUBE_SCHEDULER} ${KUBELET} ${KUBEPROXY}
do
    systemctl stop --now $i || true
    systemctl disable --now $i || true
    pkill -9 -f $i >/dev/null || true
done

# delete all containers on the host.
nerdctl --namespace k8s.io ps -a -q | xargs -I {} -n1 timeout 10 nerdctl --namespace k8s.io rm -f {}
nerdctl ps -a -q | xargs -I {} -n1 timeout 10 nerdctl --namespace k8s.io rm -f {}


# Stop containerd
systemctl stop --now $CONTAINERD || true
systemctl disable --now $CONTAINERD || true
pkill -9 $CONTAINERD >/dev/null || true

for m in $(mount | grep -E "csi/driver.longhorn.io|kubernetes.io~csi/pvc|kubernetes.io~projected/kube" | awk '{print $3}' | xargs);
do
    umount -f $m || true
done

# Unmount and clean up bindmounts we created
umount -f /var/lib/docker/kubelet
umount -f /var/lib/containerd/etcd
umount -f /var/lib/docker
umount -f /var/lib/kubelet
sed -i'' -e "/\/var\/lib\/docker\/kubelet/d" -e "/\/var\/lib\/containerd\/etcd/d" -e "/\/var\/lib\/docker/d" /etc/fstab
rm -rf /var/lib/containerd/etcd /var/lib/docker/kubelet /var/lib/kubelet

# delete all the rpms
eval ${PACKAGE_MANAGER} remove -y ${ETCD} ${KUBE_APISERVER} ${KUBE_CONTROLLER_MANAGER} ${KUBE_SCHEDULER} ${CONTAINERD} ${KUBELET} ${KUBEPROXY} ${KUBECTL} ${KUBERNETES_CNI} ${KUBEADM} ${RUNC} ${CRICTL}

# Reload systemd manager configuration
systemctl daemon-reload

# clean up following directories
rm -rf \
 /var/run/kubernetes \
 /etc/kubernetes/* \
 /etc/cni/net.d \
 /var/log/containers \
 /var/lib/containerd/* \
 /var/log/pods \
 /var/lib/etcd \
 /var/lib/cni \
 /var/lib/kube-proxy \
 /opt/ezkube/bundle/* \
 /etc/zypp/repos.d/ezkube-v* \
 /etc/yum.repos.d/ezkube-v*- \
 /root/.kube \
 /var/lib/kubelet \
 /etc/systemd/logind.conf.d/99-kubelet.conf

# Clean local repo caches
[[ "${PACKAGE_MANAGER}" == 'yum' ]] && eval ${PACKAGE_MANAGER} clean all || eval ${PACKAGE_MANAGER} clean

# Delete ezkube repo file from /etc/zypp/repo.d directory
rm -f /etc/zypp/repos.d/ezkube-*.repo
rm -f /etc/yum.repos.d/ezkube-*.repo

rm -rf /root/.kube /home/*/.kube

rm -rf /tmp/*.xtrace

# remove files from /tmp
rm -rf /tmp/ezkube-*.x86_64 /tmp/ezfab-release-*

#uninstall the agent
bash /usr/local/bin/uninstall-ezkf-agent.sh || true

rm -rf /opt/ezkf /opt/ezkube /opt/cni /opt/containerd
rm -rf /var/lib/kubelet /var/lib/calico
rm -rf /var/log/ezkf /var/log/ezkube /var/log/calico
rm -f /usr/bin/ezctl /usr/local/bin/ezkf-agent

# stop and disable keep-alived
systemctl stop keepalived && systemctl disable keepalived
