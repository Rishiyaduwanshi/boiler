import { config } from './index.js';

const isProduction = config.NODE_ENV === 'production';

export const cookieOptions = {
  accessToken: {
    httpOnly: true,
    secure: isProduction,
   sameSite: isProduction ? 'none' : 'lax',
    maxAge: isProduction ? 15 * 60 * 1000 : 2 * 60 * 60 * 1000
  },
  refreshToken: {
    httpOnly: true,
    secure: isProduction,
    sameSite: isProduction ? 'none' : 'lax',
    maxAge: isProduction ?  1000 * 60 * 60 * 24 * 7 : 1000 * 60 * 60 * 24 * 20 
  }
};
