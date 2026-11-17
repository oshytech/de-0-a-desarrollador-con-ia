package club.reservas

import androidx.compose.ui.window.ComposeUIViewController
import club.reservas.aplicacion.RelojSistema
import club.reservas.datos.RepositorioSimulado
import club.reservas.ui.App

// Punto de entrada para iOS. El proyecto de Xcode (iosApp) lo llama desde Swift.
// Para usar la API real: RepositorioHttp("http://localhost:8080"), porque el simulador
// de iOS comparte red con el Mac.
@Suppress("ktlint:standard:function-naming", "FunctionName")
fun MainViewController() = ComposeUIViewController { App(RepositorioSimulado(RelojSistema), RelojSistema, "s-1") }
