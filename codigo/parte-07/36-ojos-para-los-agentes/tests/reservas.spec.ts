import { test, expect, type Page } from "@playwright/test";

// Mañana, en formato AAAA-MM-DD y en hora local, igual que el campo de fecha.
function manana(): string {
  const d = new Date();
  d.setDate(d.getDate() + 1);
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const dia = String(d.getDate()).padStart(2, "0");
  return `${d.getFullYear()}-${m}-${dia}`;
}

// La misma fecha, como la lee el socio en la página.
function mananaLegible(): string {
  const d = new Date();
  d.setDate(d.getDate() + 1);
  return d.toLocaleDateString("es-ES", { weekday: "long", day: "numeric", month: "long" });
}

async function rellenar(page: Page, correo: string, pista: string, hora: string) {
  await page.getByLabel("Correo de socio").fill(correo);
  await page.getByLabel("Pista").selectOption({ label: pista });
  await page.getByLabel("Día").fill(manana());
  await page.getByLabel("Hora de inicio").selectOption(hora);
  await page.getByRole("button", { name: "Reservar pista" }).click();
}

test.beforeEach(async ({ page }) => {
  await page.goto("/");
});

test("una socia activa reserva una pista libre", async ({ page }) => {
  await rellenar(page, "ana@club.test", "Tenis 1", "18:00");

  await expect(page.getByRole("status")).toHaveText(
    `Reserva confirmada: Tenis 1, ${mananaLegible()} a las 18:00.`,
  );
  await expect(page.getByRole("listitem")).toHaveText([`Tenis 1, ${mananaLegible()} a las 18:00`]);
  await expect(page.getByText("Todavía no tienes reservas")).toBeHidden();
});

test("si la pista ya está ocupada, el error se ve y dice qué hacer", async ({ page }) => {
  await rellenar(page, "ana@club.test", "Pádel 1", "19:00");
  await expect(page.getByRole("status")).toBeVisible();

  await rellenar(page, "ana@club.test", "Pádel 1", "19:00");

  const alerta = page.getByRole("alert");
  await expect(alerta).toBeVisible();
  await expect(alerta).toContainText("Pádel 1 ya está reservada ese día a las 19:00");
  await expect(alerta).toContainText("Prueba otra hora u otra pista");
  await expect(page.getByRole("listitem")).toHaveCount(1);
});

test("enviar vacío marca los campos y lleva el foco al primero", async ({ page }) => {
  await page.getByRole("button", { name: "Reservar pista" }).click();

  await expect(page.getByText("Escribe el correo con el que te diste de alta")).toBeVisible();
  await expect(page.getByText("Elige el día de la reserva.")).toBeVisible();
  await expect(page.getByLabel("Correo de socio")).toBeFocused();
  await expect(page.getByLabel("Correo de socio")).toHaveAttribute("aria-invalid", "true");
});

test("el formulario se recorre entero con el teclado", async ({ page }) => {
  await page.getByLabel("Correo de socio").focus();
  // El campo de fecha de Chromium tiene tres paradas (día, mes y año),
  // así que se apunta el orden de los elementos y se quitan las repeticiones.
  const recorrido: string[] = [];
  for (let i = 0; i < 8; i++) {
    const enfocado = await page.evaluate(() => {
      const el = document.activeElement as HTMLElement;
      return el.id || el.tagName.toLowerCase();
    });
    if (recorrido.at(-1) !== enfocado) recorrido.push(enfocado);
    await page.keyboard.press("Tab");
  }
  expect(recorrido.slice(0, 5)).toEqual(["correo", "pista", "dia", "hora", "button"]);
});

test("el aspecto del formulario no cambia sin que nadie lo decida", async ({ page }) => {
  await expect(page).toHaveScreenshot("formulario-inicial.png", { fullPage: true });
});
