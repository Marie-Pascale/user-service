package httpapi

import "net/http"

// specHandler sert le fichier openapi.yaml a la racine du projet tel quel.
// Le chemin est relatif au repertoire de travail du processus (WORKDIR
// /app dans le Dockerfile), ou se trouve deja le fichier grace au volume
// monte par docker-compose.
func specHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	http.ServeFile(w, r, "openapi.yaml")
}

// docsHandler sert une page HTML minimale chargeant Swagger UI depuis un
// CDN, pointee vers /openapi.yaml. Pas besoin de vendre/embarquer les
// fichiers de Swagger UI dans le projet -- juste cette page statique.
func docsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(docsHTML))
}

const docsHTML = `<!DOCTYPE html>
<html>
<head>
  <title>Gestionnaire d'usagers -- Documentation API</title>
  <meta charset="utf-8" />
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: "/openapi.yaml",
        dom_id: "#swagger-ui",
      });
    };
  </script>
</body>
</html>`
