<h1 align="center"> NodeFitter </h1>

<p align="center"> A simple OpenNebula VM autoscaler </p>

<div align="center">
  <img alt="Static Badge" src="https://img.shields.io/badge/-1.26.5-blue?style=flat-square&logo=go&logoSize=auto&labelColor=FFF&color=00ADD8">
</div>

<div align="center">
  <img alt="Static Badge" src="https://img.shields.io/badge/-Docker-blue?style=flat-square&logo=docker&logoSize=auto&labelColor=FFF&color=2496ED">
  <img alt="Static Badge" src="https://img.shields.io/badge/-Kubernetes-blue?style=flat-square&logo=kubernetes&logoSize=auto&labelColor=FFF&color=326CE5">
</div>

## About this project
<div align="center">
<img src="./assets/img/arch(itecture).png" width="75%">
</div>

NodeFitter is a simple VM autoscaler created for the "Fog and Cloud Computing" course at <a href="https://www.unitn.it/it">University of Trento</a>, Italy.

NodeFitter is capable of automatically scale VMs running on OpenNebula. Specifically, the scaler automatically spawn one VM per template upon startup, then continues to check the internal VM memory and CPU consumption.

When the free memory or the CPU is under a configurable threshold, NodeFitter automatically creates a new VM of the same type and automatically makes the VM join the Kubernetes cluster.

The script used to obtain information about memory and CPU consumption is uploaded to the new VM, meaning that it is completely configurable. As for the Kubernetes auto-join functionality, NodeFitter automatically generate a 10-minute token and register it with the control plane.

## Installation instructions

To setup Docker and Kubernetes, as well the necessary OpenNebula templates and golden images, it is sufficient to follow the instructions reported in the [appropriate README](VMsConfig/README.md) under the <a href="./VMsConfig/">VMsConfig folder</a>.

## Configuration instruction

## Usage instructions


## Authors

- GitHub: <a href="https://github.com/LorenzoCaraffini">LorenzoCaraffini</a>
- GitHub: <a href="https://github.com/FireStoat3">FireStoat3</a>
