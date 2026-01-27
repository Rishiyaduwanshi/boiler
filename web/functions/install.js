// Cloudflare Pages Function for OS auto-detection
// Route: boiler.iamabhinav.dev/install

export async function onRequest(context) {
  const userAgent = context.request.headers.get('User-Agent') || '';
  
  // Detect OS from User-Agent
  const isWindows = userAgent.includes('Windows') || 
                    userAgent.includes('PowerShell') ||
                    userAgent.includes('WindowsPowerShell');
  
  // GitHub raw content URL
  const baseUrl = 'https://raw.githubusercontent.com/rishiyaduwanshi/boiler/main/scripts';
  const script = isWindows ? 'install.ps1' : 'install.sh';
  
  try {
    // Fetch the appropriate script
    const response = await fetch(`${baseUrl}/${script}`);
    
    if (!response.ok) {
      return new Response('Installation script not found', { status: 404 });
    }
    
    const content = await response.text();
    
    // Return with appropriate content type
    return new Response(content, {
      headers: {
        'Content-Type': 'text/plain; charset=utf-8',
        'Cache-Control': 'public, max-age=300', // 5 min cache
        'X-Detected-OS': isWindows ? 'Windows' : 'Unix',
      }
    });
  } catch (error) {
    return new Response('Error fetching installation script', { 
      status: 500,
      headers: { 'Content-Type': 'text/plain' }
    });
  }
}
