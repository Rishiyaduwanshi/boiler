import { Router } from 'express';
import appResponse from '../utils/appResponse.js';

const router = Router();

// Health check route
router.get('/health', (req, res) => {
  appResponse(res, {
    message: 'Server is healthy',
    data: { status: 'UP' }
  });
});

export default router;