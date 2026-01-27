// Cloudflare Pages Function: /install
// Auto-detect OS and serve correct install script

export async function onRequest(context) {
  const uaRaw = context.request.headers.get("user-agent") || "";
  const ua = uaRaw.toLowerCase();

  const isWindows =
    ua.includes("windows") ||
    ua.includes("powershell") ||
    ua.includes("win64") ||
    ua.includes("win32");

  // Tumhare scripts repo ke root ke /scripts folder me hain
  const baseUrl = "https://raw.githubusercontent.com/rishiyaduwanshi/boiler/main/scripts";
  const script = isWindows ? "install.ps1" : "install.sh";

  try {
    const res = await fetch(`${baseUrl}/${script}`);

    if (!res.ok) {
      return new Response("Installer script not found", { status: 404 });
    }

    const content = await res.text();

    return new Response(content, {
      headers: {
        "Content-Type": isWindows
          ? "text/plain; charset=utf-8"
          : "application/x-sh; charset=utf-8",
        "Cache-Control": "public, max-age=3600",
        "X-Detected-OS": isWindows ? "windows" : "unix",
      },
    });
  } catch (err) {
    return new Response("Failed to fetch installer", {
      status: 500,
      headers: { "Content-Type": "text/plain" },
    });
  }
}
