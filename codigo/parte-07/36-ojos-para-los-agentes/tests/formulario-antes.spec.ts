import { test, expect } from "@playwright/test";
import { pathToFileURL } from "node:url";
import { resolve } from "node:path";

// Comprueba los problemas del formulario "antes" del capítulo 35.
// Estas aserciones describen defectos: si un día fallan, es que alguien lo arregló.
const antes = pathToFileURL(resolve(__dirname, "../../35-usabilidad/formulario-antes.html")).href;

test("el formulario de antes no se deja usar sin ratón ni con lector de pantalla", async ({ page }) => {
  await page.goto(antes);

  // Sin <label>, ningún campo tiene nombre accesible.
  await expect(page.getByLabel("Correo")).toHaveCount(0);
  // El "botón" es un div: no existe para el teclado ni para un lector de pantalla.
  await expect(page.getByRole("button")).toHaveCount(0);

  // Y el error no dice qué ha fallado ni cómo arreglarlo.
  await page.locator("#hora").fill("7");
  await page.getByText("Enviar").click();
  await expect(page.locator("#mensaje")).toHaveText("Error 422: datos no válidos");
});
