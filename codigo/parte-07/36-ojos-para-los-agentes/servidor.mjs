// Servidor estático mínimo, sin dependencias. Playwright lo arranca y lo para (webServer).
import { createServer } from "node:http";
import { readFile } from "node:fs/promises";
import { extname, join, normalize } from "node:path";

const raiz = join(import.meta.dirname, "public");
// Solo en local (127.0.0.1): no se expone a la red ni despierta avisos del cortafuegos.
const host = process.env.HOST ?? "127.0.0.1";
const puerto = Number(process.env.PORT ?? 4173);
const tipos = { ".html": "text/html; charset=utf-8", ".js": "text/javascript", ".css": "text/css" };

createServer(async (peticion, respuesta) => {
  const ruta = new URL(peticion.url, "http://localhost").pathname;
  const fichero = normalize(join(raiz, ruta === "/" ? "index.html" : ruta));
  if (!fichero.startsWith(raiz)) {
    respuesta.writeHead(403).end();
    return;
  }
  try {
    const contenido = await readFile(fichero);
    respuesta.writeHead(200, { "content-type": tipos[extname(fichero)] ?? "application/octet-stream" });
    respuesta.end(contenido);
  } catch {
    respuesta.writeHead(404).end("No encontrado");
  }
}).listen(puerto, host, () => console.log(`Reservas en http://${host}:${puerto}`));
