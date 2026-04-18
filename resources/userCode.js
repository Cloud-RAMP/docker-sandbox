module.exports = {
  onMessage: (message) => {
    console.log("Got message:", message);
    return `Processed message: ${message}`;
  },

  onLeave: (userId) => {
    return `User ${userId} has been removed.`;
  },

  // Called when a user joins
  onJoin: (userId) => {
    return `Welcome, User ${userId}!`;
  },

  onError: (error) => {
    return `Error handled: ${error}`;
  },
};