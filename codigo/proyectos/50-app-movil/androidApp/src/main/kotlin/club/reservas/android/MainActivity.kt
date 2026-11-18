package club.reservas.android

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import club.reservas.aplicacion.RelojSistema
import club.reservas.datos.RepositorioHttp
import club.reservas.datos.RepositorioSimulado
import club.reservas.ui.App

// Sin inicio de sesión todavía: la app actúa como el socio s-1 (ver el capítulo 44 para la autenticación).
private const val SOCIO_DE_PRUEBA = "s-1"

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        val repositorio =
            if (BuildConfig.API_SIMULADA) {
                RepositorioSimulado(RelojSistema)
            } else {
                RepositorioHttp(BuildConfig.API_URL)
            }
        setContent { App(repositorio, RelojSistema, SOCIO_DE_PRUEBA) }
    }
}
