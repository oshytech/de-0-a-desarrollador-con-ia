// Package dominio contiene las entidades y las reglas de negocio del club:
// socios, pistas y reservas. No depende de ningún paquete externo y no sabe
// nada de bases de datos ni de HTTP.
//
// Las reglas que se comprueban aquí son las que solo necesitan la propia
// reserva y la hora actual:
//
//   - Una reserva dura una hora exacta y empieza en punto, entre las 08:00
//     y las 21:00 (la última termina a las 22:00).
//   - No se reserva en el pasado ni con más de 14 días de antelación.
//   - Una reserva se puede cancelar hasta 2 horas antes de su inicio.
//   - Solo los socios activos pueden reservar.
//
// Las reglas que necesitan consultar otras reservas (solapes en una pista y
// el máximo de reservas por socio) viven en el paquete aplicacion, que tiene
// acceso a los repositorios.
package dominio
