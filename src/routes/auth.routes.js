import { Router } from 'express';
import {
  register as signup,
  login as signin,
  refreshToken,
  logout as signout,
} from '../controllers/auth.controller.js';

const router = Router();

// Auth routes
router.post('/signup', signup);
router.post('/login', signin);
router.post('/refresh-token', refreshToken);
router.post('/logout', signout);

export default router;
