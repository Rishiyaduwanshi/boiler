import jwt from 'jsonwebtoken';
import { UnauthorizedError } from '../utils/appError.js';
import { config } from '../../config/index.js';

export const authenticate = (req, res, next) => {
  try {
    const token = req.cookies?.accessToken;

    if (!token) {
      throw new UnauthorizedError('Authentication token missing');
    }

    const decoded = jwt.verify(token, config.JWT_SECRET);
    req.user = { id: decoded.id };

    next();
  } catch (error) {
    next(new UnauthorizedError('Invalid or expired token'));
  }
};
