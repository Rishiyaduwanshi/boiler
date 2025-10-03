class AppError extends Error {
    constructor({ statusCode = 500, message = "Internal Server Error", errors = [], stack = "" }) {
      super(message);
      this.statusCode = statusCode;
      this.success = false;
      this.errors = errors;
  
      if (stack) {
        this.stack = stack;
      } else {
        Error.captureStackTrace(this, this.constructor);
      }
    }
  }
  
  export { AppError };
  


export class NotFoundError extends AppError {
  constructor(message = 'Resource not found') {
    super({ message, statusCode: 404 });
  }
}

export class BadRequestError extends AppError {
  constructor(message = 'Bad request') {
    super({ message, statusCode: 400 });
  }
}

export class UnauthorizedError extends AppError {
  constructor(message = 'Unauthorized') {
    super({ message, statusCode: 401 });
  }
}
