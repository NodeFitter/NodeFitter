# VM SETUP

## Getting memory consumption

Opennebula monitoring data include information about the hypervisor, therefore, ram consumption includes measurements of both VM AND hypervisor. The only way to get memory info about VM memory consumption is push them from inside the VM itself. The following approach uses opennebula's onegate: data will be visible under user template.

### Requirements and instructions

> [!TIP]
> To change the control plane serving address, follow these steps:
> - Create a file with the new cluster configuration named `kubeadm-config.yaml`:
>   ```yaml
>     apiVersion: kubeadm.k8s.io/v1beta4
>     kind: ClusterConfiguration
>     
>     kubernetesVersion: v1.36.3
>     
>     controlPlaneEndpoint: "192.0.2.1:6443" # Insert the new ip
>     
>     apiServer:
>       certSANs: # Insert all ip addresses that can be used to contact the control plane
>         - 192.0.2.1
>         - 192.0.2.2
>     
>   ```
> - Regenerate the certificates with:
>   ```sh
>   sudo mv /etc/kubernetes/pki/apiserver.crt /etc/kubernetes/pki/apiserver.crt.old
>   sudo mv /etc/kubernetes/pki/apiserver.key /etc/kubernetes/pki/apiserver.key.old
>   sudo kubeadm init phase certs apiserver --config kubeadm-cert-config.yaml
>   ```
> - Edit `/etc/kubernetes/manifests/kube-apiserver.yaml` to substitute any reference to the old address with the new one. Kubernetes will restart the API server upon saving;
> - Edit `/etc/kubernetes/admin.conf` to substitute any reference to the old address with the new one.

- VM template must have QEMU Guest Agent (under OS & CPU) set to yes
- VM template must have OneGate token enabled (under context)
- A script inside the vm (under `/bin`) that push the CPU and memory data with:
    ```onegate vm update --data 'TEST_KEY="test_value"```
  In this readme such file is referred as `res_info` or `res_info.sh`
  

> [!TIP]
> In case of onegate error:
>	Since the token file is in the cdrom context, mount the cdrom and add token by adding these instruction to the context:
>	```sh
>
>       mkdir -p /mnt/context
>	    mount /dev/sr0 /mnt/context
>	    export ONEGATE_TOKEN="$(cat /mnt/context/token.txt)
>
>   ```

> [!NOTE]
> The scheduler is able to automatically upload a command to generate the script to the VM. Create a script anywhere in your machine and configure its position in the scheduler configuration file (check the `config` folder). The scheduler will parse the file, and uploaded as a context variable in the new VM template. To automatically setup the script in the VM, set the following lines in the VM template start script (under context section):
> ```sh
> cat /var/run/one-context/one_env > /tmp/var
> source /tmp/var && eval $RES_SCRIPT_INSTALL_COMMAND
> ```

- Modify guestconfig.conf under `/var/lib/one/remotes/etc/im/kvm-probes.d/guestagent.conf`, setting
```
enable: true
```
and adding the command:
```
:vm_qemu_meminfo_pid: one-$vm_id '{"execute":"guest-exec","arguments":{"path":"/bin/res_info","arg":[""],"capture-output":true}}' --timeout 5
```

## Kubernetes setup

### Host instructions
- Install Docker as described here https://docs.docker.com/engine/install/ubuntu/
- Install K8 as described here https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/
- Regenerated config file of containerd (otherwise `kubeadm init` will fail due to incompatible CRI config) and enable containerd to use systemd to manage cgroups:
	```sh
    sudo containerd config default | sudo tee /etc/containerd/config.toml
    sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.t
    ```
> [!IMPORTANT]
> Disable swap with
> ```sh
> sudo swapoff -a
> ```
> To make the change permanent modify the `/etc/fstab` file by commenting out the swap.img line
- Run the following command to initialize the control plane:
    ```sh
    sudo kubeadm init
    ```
- Run
    ```sh
    mkdir -p $HOME/.kube
    sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
    sudo chown $(id -u):$(id -g) $HOME/.kube/con
    ```
- Install a CNI (for example, Calico. Control Plan VM ONLY):
  ```sh 
  kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.31.1/manifests/calico.yaml
  ```

To get a command with token to join the cluster, run
```sh
kubeadm token create --print-join-command
```
Default validity of the data given by ```kubeadm``` is 24 hours.

### Worker nodes instructions

> [!IMPORTANT]
> For guest VMs that will become worker nodes it is MANDATORY to set CPU to `westmere` and enable the following CPU features `sse4.1` `sse4.2` `sse3` `popcnt` `cx16`

- Install Docker and Kubernetes as described in the previous section. To join the cluster use the command outputted by:
  ```sh
  kubeadm token create --print-join-command
  ```
  However, the scheduler automatically inserts a valid kubeadm join command under context. Insert as start script the following commands:
  ```sh
  cat /var/run/one-context/one_env > /tmp/var
  source /tmp/var && eval $K8_JOIN_COMMAND
  ```
  If you already run the source command in the start script, adding `eval $K8_JOIN_COMMAND` is sufficient.
  If Alpine Linux is preferred, refer to https://wiki.alpinelinux.org/wiki/Docker and https://wiki.alpinelinux.org/wiki/K8s
- VM template must have SET_HOSTNAME as key and vm-$VMID as value under Context Custom Variables;
- VM has to be part of an OpenNebula VM group which name is the same as the various pods nodeselector regions;
- Pods namespaces and labels must also match VM groups names. For labels, type is used as the key.
