import logger from '../utils/errorLogger.js';
import { AppError } from '../utils/appError.js';
import { config } from '../../config/index.js';

const globalErrorHandler = (err, req, res, _) => {
  try {
    const isDev = config.NODE_ENV === 'development';
    const formattedErr =
      err instanceof AppError
        ? err
        : new AppError({
            message: isDev ? err.message : 'Something Broke!',
            statusCode: err.statusCode || 500,
            stack : err.stack || '',
            errors : isDev ? err.errors || [] : []
          });

    try {
      logger.error(
        `[${formattedErr.statusCode}] ${req.method} ${req.originalUrl} - ${req.headers['user-agent']} ${formattedErr.stack}`
      );
    } catch (logError) {
      console.error('🚨 Logger failed:', logError);
    }

    res.status(formattedErr.statusCode).json({
      message: formattedErr.message,
      statusCode: formattedErr.statusCode,
      success: formattedErr.success ?? false,
      errors: formattedErr.errors || [],
    });
  } catch (fatalError) {
    logger.error(`🔥 Critical Error in Error Handler:\n${fatalError.stack}`);
    res.status(500).json({
      message: 'Internal server fatal error',
      statusCode: 500,
      success: false,
      errors: [],
    });
  }
};

export default globalErrorHandler;
