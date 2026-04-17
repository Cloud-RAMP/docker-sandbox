const net = require('net');
const socketPath = '/tmp/cloud_ramp_socket';

// Simulate user-defined code execution
let userCode = null;

// Define handlers for each message type
const messageHandlers = {
  0: (data) => {
    // 0 = receiving initial code to execute
    console.log('Received initial code to execute');
    userCode = eval(data.toString()); // Parse the code (ensure it's trusted!)
    console.log('Code loaded successfully');
  },
  1: (data) => {
    // 1 = receiving request (coordinator -> container)
    console.log('Received request from coordinator');
    if (userCode && typeof userCode.onMessage === 'function') {
      const response = userCode.onMessage(data.toString());
      sendMessage(2, response); // Send response (type 2)
    } else {
      console.error('No user code loaded or onMessage not defined');
      sendMessage(5, 'Error: No user code loaded or onMessage not defined'); // Send error (type 5)
    }
  },
  3: (data) => {
    // 3 = receiving request (container -> coordinator)
    console.log('Received request from container');
    // Handle as needed (if applicable)
  },
  5: (data) => {
    // 5 = receiving error
    console.error(`Error received: ${data.toString()}`);
  },
};

// Function to send messages
const sendMessage = (type, payload) => {
  const message = Buffer.concat([
    Buffer.from([type]), // First byte is the message type
    Buffer.from(payload), // Remaining bytes are the payload
  ]);
  client.write(message);
};

// Create a Unix domain socket client
const client = net.createConnection(socketPath, () => {
  console.log('Connected to coordinator');
});

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