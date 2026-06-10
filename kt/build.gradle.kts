plugins {
    kotlin("jvm")
}

group = "de.arcan.arclib"
version = "0.1.0-alpha.1"

dependencies {
    testImplementation(kotlin("test"))
    testImplementation("com.google.code.gson:gson:2.11.0")
}

tasks.test {
    useJUnitPlatform()
}

kotlin {
    jvmToolchain(17)
}
