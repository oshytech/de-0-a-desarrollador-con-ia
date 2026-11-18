import SwiftUI
import UIKit
import Shared

// Envuelve la pantalla de Compose Multiplatform para que SwiftUI la muestre.
struct VistaCompose: UIViewControllerRepresentable {
    func makeUIViewController(context: Context) -> UIViewController {
        MainViewControllerKt.MainViewController()
    }

    func updateUIViewController(_ uiViewController: UIViewController, context: Context) {}
}

// Compose ya reserva el hueco de la barra de estado (el Scaffold aplica los márgenes del sistema).
// Si SwiftUI también lo hiciera, quedaría una franja vacía arriba.
struct ContentView: View {
    var body: some View {
        VistaCompose().ignoresSafeArea()
    }
}
