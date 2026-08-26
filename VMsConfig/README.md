# VM SETUP GUIDE

## Getting memory consumption

OpenNebula's monitoring data include information about the hypervisor, therefore, the provided memory consumption includes measurements of both the VM AND the hypervisor. The only way to get memory information about VM memory consumption only is by pushing them from inside the VM. The following approach uses OpenNebula's OneGate: data will be visible under the VM's user template.

> [!NOTE]
> OneGate may not work if the appropriate token (which can be activated in the VM template) is not present. In case OpenNebula fails during the installation of said token, it is possible to extract it from the CDROM context. Mount the CDROM and add the token by adding these instructions to the context:
>	```sh
> mkdir -p /mnt/context
>	mount /dev/sr0 /mnt/context
>	export ONEGATE_TOKEN="$(cat /mnt/context/token.txt)
> ```
> Keep in mind that generally this is NOT needed if the OneGate token was turned on in the template since OpenNebula configures it automatically.

## Creation of the Kubernetes cluster

> [!IMPORTANT]
> One of the Kubernetes' components, kubelet, will fail to start if swap is enabled. Additional information can be found in the [official documentation](https://kubernetes.io/docs/concepts/cluster-administration/swap-memory-management/) and on the official Kubernetes' [installation guide of kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/).
>
> Swap can be disabled by executing:
> ```sh
> sudo swapoff -a
> ```
> To make the change permanent modify the `/etc/fstab` file by commenting out the swap.img line.

- Install Docker as described here: https://docs.docker.com/engine/install/ubuntu/
- Regenerate the configuration file of containerd (otherwise `kubeadm init` will fail due to incompatible CRI configuration, specifically, a disabled plugin):
	```sh
    sudo containerd config default | sudo tee /etc/containerd/config.toml

    # Enable containerd to use systemd to manage cgroups (optional)
    sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml

    # Restart containers
    sudo systemctl restart containerd
    ```
- Install Kubernetes as described here: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/
- Run the following command to initialize the control plane:
    ```sh
    sudo kubeadm init
    ```
- Run:
    ```sh
    mkdir -p $HOME/.kube
    sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
    sudo chown $(id -u):$(id -g) $HOME/.kube/config
    ```
- Install a CNI (for example, Calico):
  ```sh 
  kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.31.1/manifests/calico.yaml
  ```

> [!NOTE]
> The autoscaler automatically inserts a valid `kubeadm join` command under context upon creating the new VM with a validity of 10 minutes. Created VMs will join the Kubernetes cluster automatically.

To get a command with a token to join the cluster, run:
```sh
kubeadm token create --print-join-command
```
The default validity of the token given by ```kubeadm``` is 24 hours.

> [!NOTE]
> The scheduler will need the Kubernetes's control plane address to execute various operations such as token creation and node removal. To change the control plane serving address, follow these steps:
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
> - Backup the existing configuration and regenerate the certificates with:
>   ```sh
>   sudo mv /etc/kubernetes/pki/apiserver.crt /etc/kubernetes/pki/apiserver.crt.old
>   sudo mv /etc/kubernetes/pki/apiserver.key /etc/kubernetes/pki/apiserver.key.old
>   sudo kubeadm init phase certs apiserver --config kubeadm-cert-config.yaml
>   ```
> - Edit `/etc/kubernetes/manifests/kube-apiserver.yaml` to substitute any reference to the old address with the new one. Kubernetes will restart the API server upon saving;
> - Edit `/etc/kubernetes/admin.conf` to substitute any reference to the old address with the new one.

## VM setup for the autoscaler

> [!IMPORTANT]
> The following guide will consider OpenNebula already installed and working. To install OpenNebula, you can read the [official guide](https://docs.opennebula.io/7.4/getting_started/install_opennebula/). It is also important to NOT have other Kubernetes installations like `minikube` in the same system.

> [!IMPORTANT]
> Additionally, in order to ensure automatic join to the cluster, ipv4 forwarding needs to be active. It can be activated by running:
> ```sh
> sudo sysctl -w net.ipv4.ip_forward=1
> ```
> To make the change permanent, modify Kubernetes ip forwarding rules under the `/etc/sysctl.d` folder or run the command as part of the OpenNebula startup script.
> Additionally, if not already configured, it is necessary to allow DNS domain translation manually by adding the following instruction to the startup script of the VM:
> ```sh
> echo "nameserver 8.8.8.8" >> /etc/resolv.conf
> ```
> **This guide will include both commands in the startup script.**

> [!TIP]
> This guide will configure Ubuntu VMs, although unofficial support for Alpine Linux exists (refer to https://wiki.alpinelinux.org/wiki/Docker and https://wiki.alpinelinux.org/wiki/K8s). Using Alpine Linux instead of Ubuntu will require changes to the resources retrieval script and to the `Docker` and `Kubernetes` installations.

> [!TIP]
> This guide will configure two templates, one for the frontend and one for the backend. The autoscaler will consider every template as a type of VM.
> To keep consistency between Kubernetes namespaces and VMs, the label that will be attached to a new VM will be equal to the name of the VM group that VM is part of: in other words, a VM of a certain VM group should be able to host pods that are part of the same namespace whose name is equal to the one of the VM group of such VM.
> To ease the management of VMs, it is suggested to keep the name of VM groups and the name of VM templates the same.

### Initial setup
- Login to OpenNebula
- Go to `Storage > Marketplaces > OpenNebula Public`, search and download the `Ubuntu Minimal 24.04` image. Wait for the image to finish downloading
- Go to `Templates > VM Groups` and create a new VM group by modifying the following parameters (leave anything else untouched or modify as you please):
  - **Name**: `frontend` (under General)
  - **Role Name**: `frontend` (under `Role Details > Role Name`. If not already present, first press `Add role` to add a new role)
- Go to `Templates > VM Groups`, create a new VM group by modifying the following parameters (leave anything else untouched or modify as you please):
  - **Name**: `backend` (under General)
  - **Role Name**: `backend` (under `Role Details > Role Name`. If not already present, first press `Add role` to add a new role)

### Frontend template setup
- Go to `Templates > VM Templates`, create a new VM template (leave anything else untouched or modify as you please):
  - **Name**: `frontend` (under `General > Information`)
  - **Logo**: `ubuntu` (under `General > Information`)
  - **VM Group**: `frontend` (under `General > Ownership > Associate VM to a VM Group`)
  - **Role**: `frontend` (under `General > Ownership > Role`)
  - **Memory**: 1024 (under `General > Memory`, minimum suggested quantity for the [sample app](https://github.com/NodeFitter/Deployment/tree/dff4b180efc3f41dc9bf082b499b7270eed53d9a/sample_app))
  - **Physical CPU**: 1 (under `General > Physical CPU`, suggested quantity)
  - **Virtual CPU**: 2 (under `General > Virtual CPU`, suggested quantity)
  - **Storage**: select `Attach disk > Image > Ubuntu Minimal 24.04 > Next > Put 20480 as Size on instantiate > Finish` (under `Advanced options > Storage`)
  - **Network**: select `Attach NIC > Enable SSH connection under Guacamole Connections > Next > Select vnet > Next > Next > Finish` (under `Advanced options > Network`)
  - **CPU Model**: select `host-passthrough` as `CPU Model` (under `Advanced options > OS & CPU > CPU Model`)
  - **QEMU Guest Agent**: select `Yes` (under `Advanced options > OS & CPU > Features`)
  - **Start Script**: copy the following start script (under `Advanced options > Context > Start script`):
    ```sh
    sysctl -w net.ipv4.ip_forward=1
    echo "nameserver 8.8.8.8" >> /etc/resolv.conf
    cat /var/run/one-context/one_env > /tmp/var
    source /tmp/var && eval $RES_SCRIPT_INSTALL_COMMAND
    eval $K8_JOIN_COMMAND
    ```
  - **Context Custom Variables**: add the following custom variables (under `Advanced options > Context > Context Custom Variables`):
    ```env
    SET_HOSTNAME=vm-$VMID
    ```

### Backend template setup
- Go to `Templates > VM Templates`, create a new VM template (leave anything else untouched or modify as you please):
  - **Name**: `backend` (under `General > Information`)
  - **Logo**: `ubuntu` (under `General > Information`)
  - **VM Group**: `backend` (under `General > Ownership > Associate VM to a VM Group`)
  - **Role**: `backend` (under `General > Ownership > Role`)
  - **Memory**: 2048 (under `General > Memory`, minimum suggested quantity for the [mariadb backend](https://github.com/NodeFitter/Deployment/tree/main/))
  - **Physical CPU**: 1 (under `General > Physical CPU`, suggested quantity)
  - **Virtual CPU**: 2 (under `General > Virtual CPU`, suggested quantity)
  - **Storage**: select `Attach disk > Image > Ubuntu Minimal 24.04 > Next > Put 20480 as Size on instantiate > Finish` (under `Advanced options > Storage`)
  - **Network**: select `Attach NIC > Enable SSH connection under Guacamole Connections > Next > Select vnet > Next > Next > Finish` (under `Advanced options > Network`)
  - **CPU Model**: select `host-passthrough` as `CPU Model` (under `Advanced options > OS & CPU > CPU Model`)
  - **QEMU Guest Agent**: select `Yes` (under `Advanced options > OS & CPU > Features`)
  - **Start Script**: copy the following start script (under `Advanced options > Context > Start script`):
    ```sh
    sysctl -w net.ipv4.ip_forward=1
    echo "nameserver 8.8.8.8" >> /etc/resolv.conf
    cat /var/run/one-context/one_env > /tmp/var
    source /tmp/var && eval $RES_SCRIPT_INSTALL_COMMAND
    eval $K8_JOIN_COMMAND
    ```
  - **Context Custom Variables**: add the following custom variables (under `Advanced options > Context > Context Custom Variables`):
    ```env
    SET_HOSTNAME=vm-$VMID
    ```

### Creation of the Golden Image
  - Go to `Instances > VMs` and create a new VM using one of the two just created templates
  - Wait for the VM to complete the startup process, connect to the VM via ssh (`onevm ssh <vm-id>`), install:
    - **Docker**: https://docs.docker.com/engine/install/ubuntu/
    - **Containerd**: 
    ```sh       
    sudo apt update

    sudo apt install containerd

    sudo mkdir /etc/containerd
    ```
    and generate and install the default configuration:
    ```sh
    sudo containerd config default | sudo tee /etc/containerd/config.toml

    # Enable containerd to use systemd to manage cgroups (optional)
    sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml

    # Restart containers
    sudo systemctl restart containerd
    ```
    - **Kubernetes**: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/
  - Shutdown the VM. Go to `Instances > VMs`, select the VM, got to the `Storage` tab, select `Save as`, name the Golden Image as `pingo`

### Final changes to templates
  - Go to `Templates > VM Templates`, modify the following elements in both of the created templates (select the template and then select `Update`):
    - **Storage**: delete the current image, then select `Attach disk > Image > pingo > Next > Put 20480 as Size on instantiate > Finish` (under `Advanced options > Storage`)
    - **Add OneGate token**: make sure to activate this option (under `Advanced options > Context > Configuration`)
    - **Report Ready to OneGate**: make sure to activate this option (under `Advanced options > Context > Configuration`)
  - Go to `Templates > VM Templates` and remove any unnecessary templates (everything except for the `frontend` and `backend` templates)

## Activation of QEMU Guest Agent

> [!NOTE]
> The autoscaler is able to automatically upload a command to copy the memory and CPU retrieval script to the VM's `bin` folder. Create a script anywhere in your machine and configure its position in the autoscaler configuration file (check the [`config`](./../config/) folder or read the main [README](../README.md)). The autoscaler will parse the file, and it will upload it to the VM as a context variable in the new VM template. The file name will be the same you choose but without the file extension: it is referred in this README as `res_info` or `res_info.sh`

In the machine where OpenNebula is installed:

- Modify `guestconfig.conf` under `/var/lib/one/remotes/etc/im/kvm-probes.d/`:
  ```sh
  # Enable the monitoring
  enabled: true
  # Add under commands
  :vm_qemu_meminfo_pid: one-$vm_id '{"execute":"guest-exec","arguments":{"path":"/bin/res_info","arg":[""],"capture-output":true}}' --timeout 5
  ```

  then apply the changes with:
  ```sh
  onehost sync --force
  ```

This will cause OpenNebula to execute the resources retrieval script every time it updates its monitoring data using the `guest-exec` functionality of the QEMU guest agent. Additional information can be found in the [QEMU wiki of Guest Agent](https://wiki.qemu.org/Features/GuestAgent), while a list of possible parameters can be found on the [QEMU Guest Agent Index](https://www.qemu.org/docs/master/qapi-qga-index.html).