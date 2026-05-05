plugins {
    kotlin("jvm") version "1.9.24"
}

group = "de.arcan.arclib"
version = "0.1.0-alpha.1"

repositories {
    mavenCentral()
}

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
