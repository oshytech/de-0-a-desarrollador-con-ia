import org.jetbrains.kotlin.gradle.dsl.JvmTarget

plugins {
    alias(libs.plugins.androidApplication)
    alias(libs.plugins.composeCompiler)
}

dependencies {
    implementation(projects.shared)
    implementation(libs.androidx.activity.compose)
    implementation(libs.compose.foundation)
    debugImplementation(libs.compose.uiTooling)
}

val apiSimulada = providers.gradleProperty("apiSimulada").getOrElse("true")
val apiUrl = providers.gradleProperty("apiUrl").getOrElse("http://10.0.2.2:8080")

android {
    namespace = "club.reservas.android"
    compileSdk =
        libs.versions.android.compileSdk
            .get()
            .toInt()

    defaultConfig {
        applicationId = "club.reservas.android"
        minSdk =
            libs.versions.android.minSdk
                .get()
                .toInt()
        targetSdk =
            libs.versions.android.targetSdk
                .get()
                .toInt()
        versionCode = 1
        versionName = "1.0.0"
        buildConfigField("boolean", "API_SIMULADA", apiSimulada)
        buildConfigField("String", "API_URL", "\"$apiUrl\"")
    }
    buildFeatures {
        buildConfig = true
    }
    buildTypes {
        getByName("release") {
            isMinifyEnabled = false
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
    }
}

kotlin {
    compilerOptions {
        jvmTarget = JvmTarget.JVM_11
    }
}
