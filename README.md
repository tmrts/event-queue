# Event Producer/Consumer Simulation
A socket server that reads and forwards events from an event source
to clients.

Clients connect through TCP and use the simple protocol described in a
section below. There are two types of clients connecting to the server:

- **One** *event source*: It will send a stream of events which may or may not require clients to be notified
- **Many** *user clients*: Each one representing a specific user, these wait for notifications for events which would be relevant to the
user they represent

## The Protocol
The protocol used by the clients is string-based (i.e. a `CRLF` control
character terminates each message). All strings are encoded in `UTF-8`.

The *event source* **connects on port 9090** and will start sending
events as soon as the connection is accepted.

The many *user clients* will **connect on port 9099**. As soon
as the connection is accepted, they will send to the server the ID of
the represented user, so that the server knows which events to
inform them of. For example, once connected a *user client* may send down:
`2932\r\n`, indicating that they are representing user 2932.

After the identification is sent, the *user client* starts waiting for
events to be sent to them. Events coming from *event source* should be
sent to relevant *user clients* exactly like read, no modification is
required or allowed.

## The Events
There are five possible events. The table below describe payloads
sent by the *event source* and what they represent:

| Payload       | Sequence #| Type         | From User Id | To User Id |
|---------------|-----------|--------------|--------------|------------|
|666\|F\|60\|50 | 666       | Follow       | 60           | 50         |
|1\|U\|12\|9    | 1         | Unfollow     | 12           | 9          |
|542532\|B      | 542532    | Broadcast    | -            | -          |
|43\|P\|32\|56  | 43        | Private Msg  | 32           | 56         |
|634\|S\|32     | 634       | Status Update| 32           | -          |

Using the verification program supplied, you will receive exactly 10000000 events,
with sequence number from 1 to 10000000. **The events will arrive out of order**.

*Note: Please do not assume that your code would only handle a finite sequence
of events, **we expect your server to handle an arbitrarily large events stream**
(i.e. you would not be able to keep all events in memory or any other storage)*

Events may generate notifications for *user clients*. **If there is a
*user client* ** connected for them, these are the users to be
informed for different event types:

* **Follow**: Only the `To User Id` should be notified
* **Unfollow**: No clients should be notified
* **Broadcast**: All connected *user clients* should be notified
* **Private Message**: Only the `To User Id` should be notified
* **Status Update**: All current followers of the `From User ID` should be notified

If there are no *user client* connected for a user, any notifications
for them will be silently ignored. *user clients* expect to be notified of
events **in the correct order**, regardless of the order in which the
*event source* sent them.

## Why Go?

- Stackless threads (goroutines), a goroutine costs around 1-2 kb memory.
- Event-based architecture under the hood (epoll, kqueue, etc.).
- Built-in communication primitives (channel, select, etc.) to synchronize communications.

## Workflow
![Workflow](./workflow.png)

There are 5 types of handlers as seen in the diagram

### Event Source Handler
Event source handler reads the information sent by the event source and sends
them to the packet handler.

### Event Consumer Handler
Event consumer handler listens the `clientListenerPort`, identifies each fresh
connected consumer (i.e. parses ID) and sends a registration request to the client registry handler.

### Client Handler
Event consumers are registered to the client.Registry and wrapped up in a goroutine.
After the registration, goroutines wait in a blocking manner to receive an event, in which
case the event.Payload is sent via the underlying communication medium.

### Client Registry Handler
Keeps a list of active/inactive client sessions and waits for an operation request.
A closure is sent to the client registry handler and the closure is executed by the handler
itself, allowing lock-less safety.

### Event Packet Handler
Event packet handler parses and processes the given event
and stores it in a hash table. When a packet with sequence number equals to the
current packet index arrives, that packet and the subsequent ones are sent to the client.Registry
where the registry handler will send notifications to the connected consumers. Packets
are purged as they are used (similar to a sliding window with one end open),
since storing each of them is infeasible.

## Challenges
Events are sent in random order and the consumers require  **in-order** delivery.
Also, since event streams are very large, meaning it is impossible to store each event.
A sliding window needs to fit into the memory to ensure the in-order delivery requirement.

However this is not a problem, because if we can't store the whole window, neither can the sender.
So, that eliminates the possibility that the sender permutes packets in a large stream and sends to us.

Sender might also generate & send events randomly. But, then the sender can't verify the results,
and if it's randomly being sent, you might as well forward them randomly without paying attention to order.
Also there is the case that it might never terminate.

The last case is a packet loss. This is a realistic scenario, so we need to consider this one.
If the communication protocol was modifiable, a simple solution would be to implement an ARQ data transfer protocol. But, that's not an option. If a packet is lost and there are no acknowledgements, the server will keep waiting for the lost packet,
which is catastrophic for the system.

We can use a greedy algorithm to determine whether or not to skip the next missing packet.
Let's say that the maximum packet No. we received is 2000, we received 500 packets and we need the 1st packet packet to proceed.

If the packet speed is 150 packets per second, and there are no duplicate packets sent (not realistic, but reasonable assumption here).
From (2000-500) / 150 = 10 seconds is the expected wait time worst-case. But, this by itself is useless. We need another
metric as well. We can use the growth rate of the maximum sequence number. For example, while we are waiting for the 1st packet
we decide to wait 10 seconds, but a second later we receive the packet numbered 20k, which make our expected wait time 100 seconds.


So, now we can decide to skip some packets to lower the expected waiting time. Also keep in mind that is algorithm is very crude.

## Running
To start the server:

```bash
go run main.go
```

To start the event source and clients:

```bash
sh followermaze.sh
```

**Note:** You can use `eventListenerPort` and `clientListenerPort` environment variables 
for configuration of both the server and the client.

### The Configuration

During development, it is possible to modify the test program behavior using the 
following environment variables:

1. **logLevel** - Default: info

   Modify to "debug" to print debug messages.

2. **eventListenerPort** - Default: 9090

   The port used by the event source.

3. **clientListenerPort** - Default: 9099

   The port used to register clients.

4. **totalEvents** - Default: 10000000

   Number of messages to send.

5. **concurrencyLevel** - Default: 100

   Number of conected users.

6. **numberOfUsers** Default: concurrencyLevel * 10
	
   Total number of users (connected or not)

7. **randomSeed** - Default: 666
	
   The seed to generate random values

8. **timeout** - Default: 20000
	
   Timeout in milliseconds for clients while waiting for new messages

9. **maxEventSourceBatchSize** - Default: 100

   The event source flushes messages in random batch sizes and ramdomize the messages
   order for each batch. For example, if this configuration is "1" the event source 
   will send only ordered messages flushing the connection for each message.

10. **logInterval** - Default: 1000

   The interval in milliseconds used to log the sent messages counter.

