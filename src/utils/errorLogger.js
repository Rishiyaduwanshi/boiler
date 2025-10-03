import { createLogger, format, transports } from 'winston';
import dayjs from 'dayjs';
import chalk from 'chalk';

const customTimestamp = format((info) => {
  info.timestamp = dayjs().format('DD-MM-YYYY hh:mm:ss A');
  return info;
});

const logger = createLogger({
  level: 'info',
  format: format.combine(
    customTimestamp(),
    format.printf(({ timestamp, level, message }) => {
      const colorLevel =
        level === 'error'
          ? chalk.red(level)
          : level === 'warn'
            ? chalk.yellow(level)
            : chalk.green(level);
      return `${timestamp} [${colorLevel}]: ${message}`;
    })
  ),
  transports: [
    new transports.File({
      filename: 'logs/error.log',
      format: format.combine(
        format.uncolorize(),
        format.printf(
          ({ timestamp, level, message }) =>
            `${timestamp} [${level}]: ${message} \n`
        )
      ),
    }),
  ],
});

export default logger;