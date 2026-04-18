const net = require('net');

// gather socket number from container arguments
const socketNumber = process.argv[2];
if (!socketNumber) {
  console.error('Error: Please provide a socket number as a command-line argument.');
  process.exit(1);
}

// construct socket path
const socketPath = `/tmp/cloud_ramp/sockets/${socketNumber}`;
let userCode = null;

// Create a Unix domain socket client
const client = net.createConnection(socketPath, () => {
  console.log('Connected to coordinator');
});

// Define handlers for each message type
const messageHandlers = {
  // received initial code to execute
  0: (data) => {
    console.log('Received initial code to execute');
    console.log(data.toString());
    userCode = eval(data.toString()); // Parse the code (ensure it's trusted!)
    console.log('Code loaded successfully');
    sendMessage(0, "");
  },

  // received a request from the coordinator
  1: (data) => {
    console.log('Received request from coordinator');
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
    return;
  },
  5: (data) => {
    // 5 = receiving error
    console.error(`Error received: ${data.toString()}`);
  },
};

// Send messages with custom protocol
const sendMessage = (type, payload) => {
  const message = Buffer.concat([
    Buffer.from([type]), // First byte is the message type
    Buffer.from(payload), // Remaining bytes are the payload
  ]);
  client.write(message);
};

// Handle incoming messages
client.on('data', (data) => {
  const messageType = data[0]; // First byte is the message type
  const payload = data.slice(1); // Remaining bytes are the payload

  if (messageHandlers[messageType]) {
    messageHandlers[messageType](payload);
  } else {
    console.error(`Unknown message type: ${messageType}`);
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