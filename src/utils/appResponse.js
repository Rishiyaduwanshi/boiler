export default function appResponse(
  res,
  { statusCode = 200, message = 'Success', success = true, data = [] }
) {
  res.status(statusCode).json({
    message,
    statusCode,
    success,
    data: data ?? [],
  });
}
