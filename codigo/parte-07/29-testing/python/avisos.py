"""Envía recordatorios de reserva a través de una API de mensajería externa."""


class Notificador:
    """Pide el token una vez y reutiliza el mismo cliente para todos los envíos."""

    def __init__(self, transporte, usuario, clave):
        self._transporte = transporte
        self._usuario = usuario
        self._clave = clave
        self._token = None

    def _cabeceras(self):
        if self._token is None:
            respuesta = self._transporte.post("/token", {"usuario": self._usuario, "clave": self._clave})
            self._token = respuesta["token"]
        return {"Authorization": f"Bearer {self._token}"}

    def recordar(self, correo, pista, hora):
        texto = f"Te esperamos en la {pista} a las {hora}."
        self._transporte.post("/mensajes", {"para": correo, "texto": texto}, cabeceras=self._cabeceras())
