# BrainLocal Technical Design

## 1. Overview

BrainLocal is an in-memory implementation of the Brain interface in the rModel framework. It provides a fully in-memory running Brain instance, suitable for local environment Brain operations without requiring additional storage or distributed system support.

## 2. Core Modules

### 2.1 BrainLocal Struct

The BrainLocal struct is the implementation of the Brain interface, which includes the following main fields:

- **id**: The unique identifier of the Brain.
- **labels**: Labels of the Brain.
- **neurons**: An index of Neurons, storing the mapping of all Neurons.
- **links**: An index of Links, storing the mapping of all Links.
- **state**: The current state of the Brain.
- **mu**: Read-write lock for the Brain's state.
- **cond**: Used to determine whether the Brain is in the expected state, implementing the `Wait()` method of Brain.

### 2.2 Neuron Struct

A Neuron represents a computational unit, and mainly contains the following:

- **id**: The unique identifier of the Neuron.
- **processor**: The processing logic.
- **inLinks**: Incoming Links.
- **outLinks**: Outgoing Links.
- **triggerGroups**: Trigger groups.
- **castGroups**: Propagation groups.

### 2.3 Link Struct

A Link represents the connection between Neurons and includes the following:

- **id**: The unique identifier of the Link.
- **spec**: The connection specification (source and target Neurons).
- **status**: The connection status.

### 2.4 Brain Memory

BrainMemory is the context implementation of the Brain, using the [Ristretto](https://github.com/dgraph-io/ristretto) cache library for efficient context management:

- **cache**: Ristretto cache instance.
- **numCounters**: The number of keys used for tracking frequency.
- **maxCost**: The maximum cost of the cache.

### 2.5 Brain Maintainer

BrainMaintainer is responsible for managing the Brain's runtime state. It uses channels to manage various events that drive the Brain’s operation:

- **bQueue**: The channel used for processing Brain events.
- **stop**: The channel used to stop the Brain.
- **NeuronRunner**: Responsible for concurrent execution of Neurons.

### 2.6 NeuronRunner

NeuronRunner is a part of BrainMaintainer, focusing on managing the concurrent execution of Neurons:

- **nQueue**: Neuron execution queue.
- **nQueueLen**: The length of the queue.
- **nWorkerNum**: The number of worker threads.

## 3. Main Workflow

### 3.1 Brain Construction

1. The BrainLocal instance is created through the `BuildBrain` function.
2. Neurons and Links are created according to the provided Blueprint.
3. Initial configurations (such as logging, number of worker threads, etc.) are set.

### 3.2 Brain Execution

1. The Brain execution is triggered through the `Entry` or `TrigLinks` methods.
2. Based on the triggered Links, the corresponding Neurons are activated.
3. Neurons execute their processing logic and may read/write to the Memory.
4. Based on the output of the Neurons and the configuration of Links, downstream Neurons are activated.

## 4. Concurrency Control

- Mutexes and condition variables are used to ensure thread safety for Brain operations.
- Support for concurrent execution of multiple Neurons.
- A `Wait` method is provided to wait for the Brain to complete execution.

## 5. Performance Considerations

- Ristretto cache is used to improve memory read/write speed.
- The number of worker threads can be configured to balance resource usage and concurrency.

## 6. Future Optimization Directions

- Enhance monitoring and debugging functionality for easier problem diagnosis.
- Optimize memory management strategies to improve large-scale data processing capabilities.
