process.on('uncaughtException', (err) => {
  console.error('UNCAUGHT EXCEPTION! Shutting down...');
  console.error(err.name, err.message);
  console.error(err);
  process.exit(1);
});

import { config } from './config/index.js';
import app from './src/app.js';
import dayjs from 'dayjs';
import './db/connectDb.js'
const PORT = config.PORT || 5440;

const server = app.listen(PORT, () => {
  const now = dayjs().format('DD-MM-YYYY HH:mm:ss A');
  const cyan = '\x1b[36m';
  const green = '\x1b[32m';
  const yellow = '\x1b[33m';
  const reset = '\x1b[0m';
  console.log(
    `${yellow}[${now}]${reset} ${green}Server is running${reset} in ${cyan}${config.NODE_ENV}${reset} mode at ${cyan}${config.APP_URL}${reset}`
  );
});

process.on('unhandledRejection', (err) => {
  console.error('UNHANDLED REJECTION! Shutting down...');
  console.error(err.name, err.message);
  console.error(err);
  server.close(() => {
    process.exit(1);
  });
});

process.on('SIGINT', () => {
  console.log('SIGINT received. Shutting down gracefully...');
  server.close(() => {
    console.log('Process terminated with Ctrl+C!');
  });
});

process.on('SIGTERM', () => {
  console.log('SIGTERM received. Shutting down gracefully...');
  server.close(() => {
    console.log('Process terminated!');
  });
});
