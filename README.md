# PacketCapture Controller (Poor-Man’s Antrea PacketCapture)-LFX-Task-Term-01-2026

This project implements a simplified version of **Antrea PacketCapture** as part of the LFX evaluation task.  
It demonstrates Kubernetes controller fundamentals, DaemonSets, Pod annotations, and node-level packet capture using `tcpdump`.

---

## 📌 Problem Statement

The goal is to build a Kubernetes controller that:

- Runs as a **DaemonSet** (one Pod per node)
- Watches Pods running on the same node
- Starts packet capture when a specific **annotation** is added to a Pod
- Stops packet capture and cleans up files when the annotation is removed

---

## 🧠 Key Idea (In Simple Terms)

- Pods can be annotated with extra metadata
- When a Pod gets annotated with `tcpdump.antrea.io: "<N>"`
- The controller running on the **same node** starts `tcpdump`
- Captured packets are saved as `.pcap` files
- Removing the annotation stops capture and deletes files

---

## 🏗️ Architecture

![Architecture Diagram](architecture.png)

### Components

- **Kind Cluster** – local Kubernetes cluster
- **Antrea CNI** – networking layer
- **PacketCapture DaemonSet**
  - Runs on every node
  - Contains a Go controller + tcpdump
- **Application Pod**
  - Generates network traffic
  - Annotated to trigger capture

---

## 🔄 Control Flow

1. User adds annotation to a Pod  
2. Controller detects the annotation via Pod watch  
3. Controller starts tcpdump on that node  
4. tcpdump writes rotating `.pcap` files  
5. Annotation is removed  
6. Controller stops tcpdump and deletes files  

---

## 📎 Annotation Format

```yaml
tcpdump.antrea.io: "3"
