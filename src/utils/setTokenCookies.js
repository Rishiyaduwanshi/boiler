import { cookieOptions } from '../../config/cookie.js';

const setTokenCookies = (res, tokens) => {
  res.cookie('accessToken', tokens.accessToken, cookieOptions.accessToken);
  res.cookie('refreshToken', tokens.refreshToken, cookieOptions.refreshToken);
};

export default setTokenCookies;
