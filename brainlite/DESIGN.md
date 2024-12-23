# BrainLite Technical Design Document

## 1. Overview

BrainLite is a lightweight implementation of the Brain interface in the rModel framework. It introduces some modifications on top of BrainLocal, with the primary difference being the implementation of BrainMemory.

## 2. Core Modules

The structures for Brain, Neuron, and Link in BrainLite are essentially the same as those in BrainLocal, so they will not be elaborated here. The main difference lies in the implementation of BrainMemory.

### 2.1 Brain Memory

BrainLite's BrainMemory is implemented using an SQLite database:

- **db**: SQLite database connection
- **datasourceName**: Database file name, default is `${brain_id}.db`
- **keepMemory**: Whether to retain the database file after Brain Shutdown

Compared to the in-memory context implementation in BrainLocal, this approach has the following features:

- **Persistent storage support**: It allows the context to be restored after Brain restart.
- **Handling larger data**: It is not limited by memory, making it suitable for larger-scale data processing.
- **Multi-language Processor support**: The SQLite-based BrainMemory enables multiple processors in different programming languages to read and write together.

### 2.2 Brain Maintainer

BrainMaintainer is one of the core components of BrainLite. Currently, it is implemented similarly to BrainLocal. However, it is planned to be refactored in the future to support multi-language BrainContext implementations.

## 3. Future Optimization Directions

- **Support for multi-language processors**: Future versions plan to support processors implemented in different programming languages, enhancing the system’s flexibility and scalability.
