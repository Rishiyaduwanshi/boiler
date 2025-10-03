import { BadRequestError, UnauthorizedError } from '../utils/appError.js';
import appResponse from '../utils/appResponse.js';
import UserModel from '../models/user.model.js';
import setTokenCookies from '../utils/setTokenCookies.js';
import jwt from 'jsonwebtoken';
import { config } from '../../config/index.js';

// Generate tokens
const generateTokens = (userId) => {
  const accessToken = jwt.sign({ id: userId }, config.JWT_SECRET, {
    expiresIn: config.JWT_EXPIRY,
  });

  const refreshToken = jwt.sign({ id: userId }, config.JWT_REFRESH_SECRET, {
    expiresIn: config.JWT_REFRESH_EXPIRY,
  });

  return { accessToken, refreshToken };
};

// Register
export const register = async (req, res, next) => {
  try {
    const { email, password, name } = req.body;

    if (!email || !password) {
      throw new BadRequestError('Email and password are required');
    }

    const existingUser = await UserModel.findOne({ email });
    if (existingUser) {
      throw new BadRequestError('User with this email already exists');
    }

    // Use email as default name if name is not provided
    const userName = name || email.split('@')[0];

    const user = await UserModel.create({ email, password, name: userName });
    const tokens = generateTokens(user._id);
    await UserModel.updateRefreshToken(user._id, tokens.refreshToken);

    const userResponse = user.toObject();
    delete userResponse.password;
    delete userResponse.refreshToken;

    setTokenCookies(res, tokens);

    appResponse(res, {
      statusCode: 201,
      message: 'User registered successfully',
      data: { user: userResponse, token: tokens.accessToken },
    });
  } catch (error) {
    next(error);
  }
};

// Login
export const login = async (req, res, next) => {
  try {
    const { email, password } = req.body;

    if (!email || !password) {
      throw new BadRequestError('Email and password are required');
    }

    const user = await UserModel.findOne({ email });
    if (!user || !(await user.comparePassword(password))) {
      throw new UnauthorizedError('Invalid credentials');
    }

    const tokens = generateTokens(user._id);
    await UserModel.updateRefreshToken(user._id, tokens.refreshToken);

    const userResponse = user.toObject();
    delete userResponse.password;
    delete userResponse.refreshToken;

    setTokenCookies(res, tokens);

    appResponse(res, {
      message: 'Login successful',
      data: { user: userResponse, token: tokens.accessToken },
    });
  } catch (error) {
    next(error);
  }
};

// Refresh Token
export const refreshToken = async (req, res, next) => {
  try {
    const { refreshToken } = req.cookies;
    if (!refreshToken) {
      throw new UnauthorizedError('Refresh token not found in cookie');
    }

    const decoded = jwt.verify(refreshToken, config.JWT_REFRESH_SECRET);
    const user = await UserModel.findById(decoded.id);
    if (!user || user.refreshToken !== refreshToken) {
      throw new UnauthorizedError('Invalid refresh token');
    }

    const tokens = generateTokens(user._id);
    await UserModel.updateRefreshToken(user._id, tokens.refreshToken);

    setTokenCookies(res, tokens);

    appResponse(res, {
      message: 'Token refreshed successfully',
    });
  } catch (error) {
    if (error instanceof jwt.JsonWebTokenError) {
      next(new UnauthorizedError('Invalid refresh token'));
    } else {
      next(error);
    }
  }
};

// Logout
export const logout = async (req, res, next) => {
  try {
    const userId = req.user?.id;
    if (userId) {
      await UserModel.updateRefreshToken(userId, null);
    }

    res.clearCookie('accessToken');
    res.clearCookie('refreshToken');

    appResponse(res, {
      message: 'Logged out successfully',
    });
  } catch (error) {
    next(error);
  }
};