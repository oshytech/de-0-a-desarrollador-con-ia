# Flujo de una reserva

Qué pasa desde que un socio pide una pista hasta que recibe la respuesta. Las comprobaciones del dominio están en `internal/dominio`; las que consultan otras reservas, en el caso de uso `ReservarPista`.

```mermaid
flowchart TD
    A[Socio pide pista y hora] --> B{¿Socio activo?}
    B -- No --> X1[ErrSocioInactivo]
    B -- Sí --> C{¿En punto, de 08:00 a 21:00?}
    C -- No --> X2[ErrHorarioNoValido]
    C -- Sí --> D{¿Futura y a 14 días o menos?}
    D -- No --> X3[ErrFueraDePlazo]
    D -- Sí --> E{¿Pista libre a esa hora?}
    E -- No --> X4[ErrPistaOcupada]
    E -- Sí --> F{¿Menos de 2 reservas futuras?}
    F -- No --> X5[ErrLimiteReservas]
    F -- Sí --> G[Reserva confirmada]
```
