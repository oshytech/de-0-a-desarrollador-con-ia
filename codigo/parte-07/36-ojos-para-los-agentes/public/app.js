// Página de reservas del club, sin frameworks. Los datos viven en memoria:
// es una demo para que un navegador (y un agente) tenga algo que mirar.

const socios = {
  "ana@club.test": { nombre: "Ana", activo: true },
  "luis@club.test": { nombre: "Luis", activo: false },
};
const nombresPista = { "tenis-1": "Tenis 1", "tenis-2": "Tenis 2", "padel-1": "Pádel 1" };
const reservas = []; // { correo, pista, dia, hora }

const form = document.getElementById("form-reserva");
const correo = document.getElementById("correo");
const dia = document.getElementById("dia");
const hora = document.getElementById("hora");
const boton = form.querySelector("button");
const resultado = document.getElementById("resultado");

// Regla 1: las reservas empiezan en punto entre las 08:00 y las 21:00.
// Si el desplegable solo ofrece horas válidas, el error no puede ocurrir.
for (let h = 8; h <= 21; h++) {
  const valor = String(h).padStart(2, "0") + ":00";
  hora.add(new Option(valor, valor));
}

// Regla 6: ni en el pasado ni con más de 14 días de antelación.
function fechaLocal(fecha) {
  const m = String(fecha.getMonth() + 1).padStart(2, "0");
  const d = String(fecha.getDate()).padStart(2, "0");
  return `${fecha.getFullYear()}-${m}-${d}`;
}
const hoy = new Date();
dia.min = fechaLocal(hoy);
dia.max = fechaLocal(new Date(hoy.getFullYear(), hoy.getMonth(), hoy.getDate() + 14));

// "2026-11-24" -> "martes, 24 de noviembre": el socio no habla en formato ISO.
function fechaLegible(valor) {
  const [a, m, d] = valor.split("-").map(Number);
  return new Date(a, m - 1, d).toLocaleDateString("es-ES", { weekday: "long", day: "numeric", month: "long" });
}

function mostrarError(campo, texto) {
  const p = document.getElementById("error-" + campo.id);
  p.textContent = texto;
  p.hidden = false;
  campo.setAttribute("aria-invalid", "true");
}

function limpiarError(campo) {
  const p = document.getElementById("error-" + campo.id);
  p.textContent = "";
  p.hidden = true;
  campo.removeAttribute("aria-invalid");
}

function validarCorreo() {
  const valor = correo.value.trim();
  if (valor === "") return "Escribe el correo con el que te diste de alta en el club.";
  if (!valor.includes("@")) return "El correo tiene que incluir @, por ejemplo nombre@ejemplo.com.";
  return "";
}

function validarDia() {
  if (dia.value === "") return "Elige el día de la reserva.";
  if (dia.value < dia.min || dia.value > dia.max) {
    return "Solo se puede reservar desde hoy hasta dentro de 14 días.";
  }
  return "";
}

// Validar al salir del campo, no mientras se escribe.
correo.addEventListener("blur", () => {
  const error = correo.value ? validarCorreo() : "";
  error ? mostrarError(correo, error) : limpiarError(correo);
});
correo.addEventListener("input", () => limpiarError(correo));
dia.addEventListener("change", () => limpiarError(dia));

function avisar(texto, tipo) {
  resultado.innerHTML = "";
  const p = document.createElement("p");
  p.className = tipo;
  p.setAttribute("role", tipo === "exito" ? "status" : "alert");
  p.textContent = texto;
  resultado.append(p);
}

function pintarReservas() {
  const lista = document.getElementById("reservas");
  lista.innerHTML = "";
  for (const r of reservas) {
    const li = document.createElement("li");
    li.textContent = `${nombresPista[r.pista]}, ${fechaLegible(r.dia)} a las ${r.hora}`;
    lista.append(li);
  }
  document.getElementById("sin-reservas").hidden = reservas.length > 0;
}

function reservar(datos) {
  const socio = socios[datos.correo];
  if (!socio) {
    return "No encontramos ese correo entre los socios. Revisa que esté bien escrito.";
  }
  if (!socio.activo) {
    return "Tu ficha de socio está inactiva. Pasa por recepción para reactivarla.";
  }
  const ocupada = reservas.some(
    (r) => r.pista === datos.pista && r.dia === datos.dia && r.hora === datos.hora,
  );
  if (ocupada) {
    return `${nombresPista[datos.pista]} ya está reservada ese día a las ${datos.hora}. ` +
      "Prueba otra hora u otra pista.";
  }
  const suyas = reservas.filter((r) => r.correo === datos.correo).length;
  if (suyas >= 2) {
    return "Ya tienes 2 reservas pendientes, que es el máximo. Cancela una para reservar otra.";
  }
  reservas.push(datos);
  return "";
}

form.addEventListener("submit", async (evento) => {
  evento.preventDefault();
  const errorCorreo = validarCorreo();
  const errorDia = validarDia();
  errorCorreo ? mostrarError(correo, errorCorreo) : limpiarError(correo);
  errorDia ? mostrarError(dia, errorDia) : limpiarError(dia);
  if (errorCorreo || errorDia) {
    (errorCorreo ? correo : dia).focus(); // llevar al usuario al primer campo con error
    return;
  }

  // Estado visible mientras "llega" la respuesta, y sin doble envío.
  boton.disabled = true;
  boton.textContent = "Reservando…";
  await new Promise((r) => setTimeout(r, 300));

  const datos = { correo: correo.value.trim(), pista: form.pista.value, dia: dia.value, hora: hora.value };
  const error = reservar(datos);
  if (error) {
    avisar(error, "error");
  } else {
    avisar(`Reserva confirmada: ${nombresPista[datos.pista]}, ${fechaLegible(datos.dia)} a las ${datos.hora}.`, "exito");
    pintarReservas();
  }
  boton.disabled = false;
  boton.textContent = "Reservar pista";
});

pintarReservas();
