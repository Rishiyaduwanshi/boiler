import morgan from 'morgan';
import fs from 'fs';
import path from 'path';
import dayjs from 'dayjs';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const logDir = path.resolve('logs');
if (!fs.existsSync(logDir)) {
  fs.mkdirSync(logDir, { recursive: true });
}

const accessLogStream = fs.createWriteStream(path.join(logDir, 'app.log'), {
  flags: 'a',
});

morgan.token('ist-date', () => {
  return dayjs().format('DD-MM-YYYY hh:mm:ss A');
});

morgan.token('user', (req) => {
  return req.user ? req.user.id : 'Guest'; 
});

morgan.token('ip', (req) => {
  return req.ip || req.headers['x-forwarded-for'] || req.connection.remoteAddress;
});

const customMorganFormat =
  '[:ist-date] :method :url :status :response-time ms - :res[content-length] :ip :user';

const httpLogger = morgan(customMorganFormat, { stream: accessLogStream });

export default httpLogger;
