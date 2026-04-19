const net = require('net');

const host = 'host.docker.internal';
const port = process.argv[2];
if (!port) {
  console.error("Failed to read port argument");
  process.exit(1);
}
let userCode = null;

// Create a TCP client
const client = net.createConnection({ host, port: port }, () => {
  console.log('Connected to coordinator');
});

// Define handlers for each message type
const messageHandlers = {

  // received initial code to execute
  0: (data) => {
    userCode = eval(data.toString());
    console.log("Received following code:\n", data.toString());
    sendMessage(0, "");
  },

  // received a request from the coordinator
  1: (data) => {
    if (userCode && typeof userCode.onMessage === 'function') {
      const response = userCode.onMessage(data.toString());
      sendMessage(3, "Some request"); // simulate external request
    } else {
      console.error('No user code loaded or onMessage not defined');
      sendMessage(5, 'Error: No user code loaded or onMessage not defined'); // Send error (type 5)
    }
  },
  4: (data) => {
    // Response from coordinator -> container
    // for simulation purposes, do nothing here
    sendMessage(6, "Done!");
  },
  5: (data) => {
    // 5 = receiving error
    console.error(`Error received: ${data.toString()}`);
  },
};

// Send messages with custom protocol
const sendMessage = (type, payload) => {
  const payloadBuffer = Buffer.from(payload);
  const length = payloadBuffer.length;
  
  const message = Buffer.concat([
    Buffer.from([type]),                           // 1 byte: message type
    Buffer.alloc(4),                               // 4 bytes: length (will fill next)
  ]);
  message.writeUInt32BE(length, 1);                // Write length at offset 1
  
  const fullMessage = Buffer.concat([message, payloadBuffer]);
  client.write(fullMessage);
};

let buffer = Buffer.alloc(0);

client.on('data', (data) => {
  buffer = Buffer.concat([buffer, data]);
  
  // Process complete messages
  while (buffer.length >= 5) {
    const messageType = buffer[0];
    const length = buffer.readUInt32BE(1);
    
    if (buffer.length < 5 + length) {
      // Not enough data yet
      break;
    }
    
    const payload = buffer.slice(5, 5 + length);
    
    if (messageHandlers[messageType]) {
      messageHandlers[messageType](payload);
    } else {
      console.error(`Unknown message type: ${messageType}`);
    }
    
    buffer = buffer.slice(5 + length);
  }
});

// Handle connection close
client.on('end', () => {
  console.log('Disconnected from coordinator');
});

// Handle errors
client.on('error', (err) => {
  console.error(`Error: ${err.message}`);
});

client.on('close', (hadError) => {
    console.log(`Socket fully closed. Error: ${hadError}`);
});