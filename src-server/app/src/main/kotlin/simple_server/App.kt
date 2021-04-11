package simple_server

import io.javalin.Javalin

object App {
    fun serve(port: Int) {
        val hasPublic = this.javaClass.classLoader
            .getResource("public") != null

        val app = Javalin.create {
            it.showJavalinBanner = false

            if (System.getenv("DEBUG") != null) {
                it.enableDevLogging()
            }

            if (hasPublic) {
                it.addStaticFiles("/public")
                it.addSinglePageRoot("/", "/public/index.html")
            }
        }

        if (!hasPublic) {
            app.get("/") { ctx -> ctx.redirect("http://example.com") }
        }

        app.start(port)
    }
}

fun main() {
    App.serve(System.getenv("PORT")?.toInt() ?: 22979)
}
