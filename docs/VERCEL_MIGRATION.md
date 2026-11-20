# Portage Vercel d’Ech0

État du chantier au 20 novembre 2026. La référence est `a9445bb44e762ca9101cae51ea1d31b070114e9c`, commune au fork et au HEAD amont lors de l’audit. Ce document distingue la faisabilité du résultat des tests. La branche ne constitue pas encore une distribution Vercel de production.

## Application actuelle et cible

`cmd/ech0/main.go` initialise le processus avant Cobra. `serve` construit le graphe Wire, amorce la configuration métier, lance le gestionnaire de travaux, le planificateur et le serveur TCP. Gin distribue les routes Huma et les routes brutes aux mêmes services Go. Le binaire inclut `template/dist`, produit par Vite depuis `web` avec Node 26 et pnpm 12.16.1. Le module Go demande 1.27.4. Les applications `hub` et `site` ont leurs propres builds.

La cible conserve la même origine HTTPS et le routeur Gin. Les assets immuables peuvent être servis par le CDN. Les pages dont les métadonnées sont rendues par Go, la découverte, OAuth, MCP et les routes API doivent passer par Go avant toute règle SPA. Un handler HTTP réutilise les dépendances de son instance sans ouvrir de port. Les données partagées vont dans libSQL, les octets dans un bucket S3 privé, les fichiers de travail bornés dans le répertoire temporaire. Le démarrage doit échouer explicitement si le stockage distant est indisponible. [Runtime Go](https://vercel.com/docs/functions/runtimes/go)

## Matrice de compatibilité

« Compatible » décrit un protocole ou composant réutilisable une fois ses dépendances disponibles. « Adaptation » indique du code ou une validation spécifique. « Bloqué » désigne une garantie impossible sur Hobby seul dans les conditions indiquées.

| Fonction et emplacement | État cible | Travail et critère d’acceptation |
|---|---|---|
| Vue, thèmes, langues, éditeur, extensions, galerie, routes SPA (`web`) | Compatible | Comparer liens profonds, métadonnées, assets et cache avec le binaire. |
| Timeline, tags, favoris/réactions, utilisateurs, commentaires, réglages, initialisation (`internal/router`) | Adaptation | Préserver relations, visibilité, pagination, transactions et premier administrateur unique sous concurrence. |
| Recherche texte | Compatible | Les repositories echo, file et comment utilisent LIKE paramétré. Aucun index FTS5 identifié dans ces chemins. Comparer accents, caractères génériques et tri. |
| GORM et migrations (`internal/database`) | Adaptation | Pilote distant database/sql, pool borné, validation de version du schéma. Isoler vec et les options SQLite natives au build. |
| Embeddings et recherche sémantique (`internal/repository/embedding`) | Adaptation | vec0 et MATCH utilisent les vecteurs float32. Tester une table ordinaire et vector_distance_l2 en recherche exacte, la dimension, le suréchantillonnage, les suppressions et le filtre d’auteur. Un index approximatif changerait le classement. [Vecteurs libSQL](https://docs.turso.tech/features/ai-and-embeddings) |
| Copilot, résumés, RAG, chat SSE | Adaptation | Préserver fournisseurs, outils, filtrage et ordre des événements. Tester annulation, premier octet, durée et reconnexion. |
| Sessions, JWT, jetons d’accès, révocation, OAuth social/OIDC, passkeys, device flow, serveur OAuth MCP | Adaptation | Secret partagé, origine canonique, challenges et codes temporaires consommés atomiquement avec expiration. Tester le début et la fin des flux sur deux instances. |
| Captcha, jeton formulaire, limitation de débit, visiteurs (`internal/middleware`, `internal/captcha`, `internal/visitor`) | Adaptation | Déplacer compteurs et consommations uniques vers un stockage partagé. Préserver le chemin intégration et ses scopes. |
| RSS, Connect/fédération, découverte des agents, MCP JSON-RPC, OpenAPI | Compatible | Préserver types de contenu, URL externe, règles d’egress, jetons, audiences, scopes et 404 sous /.well-known. |
| Fichiers, images, vidéos, audio, avatars, pièces jointes | Adaptation | VireFS sait signer un PUT. Le service actuel exige un admin, signe pour 24 h puis crée File et TempFile avant réception. Introduire staging privé, courte durée, contrôle réel de taille et type, promotion vers une clé finale immuable et finalisation idempotente par le propriétaire. Ne jamais conserver une URL signée expirante en DB. |
| Gros POST multipart des clients existants | Bloqué | Vercel rejette au-delà de 4,5 MB avant Go. Le navigateur peut transférer directement vers R2. Un client qui exige le POST historique de grande taille requiert une entrée HTTP hors de cette limite. [Limites des Functions](https://vercel.com/docs/functions/limitations) |
| Jobs, Busen, webhooks, indexation | Adaptation | Insérer l’événement avec la mutation métier. Réserver le travail par écriture conditionnelle, bail expirant et identifiant de tentative. Persister les étapes et livraisons. Un timeout après réception d’un webhook impose une clé stable pour la déduplication côté destinataire. |
| Nettoyage et statistiques | Adaptation | Nettoyage toutes les 72 h et flush visiteurs toutes les 60 min dans internal/task/scheduled. Écrire des compteurs durables à la visite et planifier le nettoyage par échéance. |
| Snapshots à cron configurable même sans trafic | Bloqué | Hobby accepte un cron quotidien avec précision horaire. Un réveil quotidien peut traiter les échéances quotidiennes. Un rattrapage tardif ne constitue pas un snapshot historique. Pour une cadence plus fine, étudier un déclencheur externe avec accord. [Cron Hobby](https://vercel.com/docs/cron-jobs/usage-and-pricing) |
| Export ZIP, téléchargement, import Ech0/Memos, Capsule et destinations fs/S3 | Adaptation | Les exporters appellent VACUUM INTO sur un fichier local avant l’archivage. Produire une vue cohérente du primaire distant et un manifeste des objets. Importer dans une génération isolée, valider puis basculer atomiquement. Tester avec médias locaux et S3. |
| Dashboard, logs historiques, live WS et SSE | Adaptation | Le dashboard interroge un fichier et s’abonne aux logs du processus. Persister le flux filtré avec rétention. Tester séparément l’upgrade Gorilla sur le runtime Go et les reconnexions multi-instances. La disponibilité des WebSockets Vercel sur tous les plans ne valide pas cet adaptateur. [WebSockets](https://vercel.com/docs/functions/websockets) |
| CLI, TUI, Capsule CLI, Docker, sites satellites | Compatible | Conserver les entrées natives et le volume data. Le TUI interactif s’exécute sur le poste de l’opérateur. Comparer les exports CLI avec le format admin. |

## Données, sécurité et exploitation

Le DSN local active WAL, busy_timeout=5000, synchronous=NORMAL et txlock=immediate. La configuration ne force pas les clés étrangères dans ce DSN. AutoMigrate précède quatorze migrateurs applicatifs. Leurs marqueurs sont lus puis écrits séparément. WithStopOnError arrête la boucle sans transmettre l’échec à son appelant. La reconstruction des colonnes temporelles consulte sqlite_master et PRAGMA table_info, recrée index et triggers et tente de changer foreign_keys dans la transaction. Ne pas lancer cette séquence sur plusieurs démarrages à froid. Préparer les versions de schéma hors requête avec verrou durable, enregistrer le succès dans la même transaction que les changements et vérifier foreign_key_check.

Pour migrer une installation, travailler d’abord sur un snapshot cohérent de test, migrer sa copie avec le binaire de la même révision, inventorier les fichiers par clé et empreinte et vérifier les références privées. Exporter schéma et lignes ordinaires vers une base libSQL neuve. Extraire explicitement les vecteurs sous sqlite-vec (ne pas importer ses tables internes) puis les convertir et comparer les voisins. Réconcilier nombres de lignes, relations, index, empreintes et ACL. Répéter un aller-retour vers une installation SQLite vierge avant la bascule avec suspension des écritures. Garder le snapshot source et définir comment rejouer les écritures postérieures avant tout retour arrière. Ce parcours doit encore être automatisé et éprouvé.

Le driver officiellement indiqué pour le moteur libSQL distant est libsql-client-go. Le pilote tursogo-serverless cible un autre moteur. Séparer URL et jeton, exiger TLS en production et éviter d’inclure les identifiants dans les erreurs. [Référence Go Turso](https://docs.turso.tech/sdk/go/reference)

Préparer un bucket R2 Standard privé, une clé S3 limitée à ce bucket et un endpoint de compte. Autoriser uniquement l’origine HTTPS de l’application dans CORS, les méthodes et les en-têtes signés réellement nécessaires. Une URL PUT peut être réutilisée jusqu’à expiration. Valider et copier une version déterminée avant publication, purger les objets temporaires expirés. Pour les téléchargements, résoudre l’autorisation de lecture à chaque émission de GET signé. Conserver l’URL applicative stable pour les médias publics. [URLs présignées R2](https://developers.cloudflare.com/r2/api/s3/presigned-urls/)

Conserver le filtrage SSRF de pkg/egress pour Connect et les webhooks, l’échappement Markdown côté rendu, les restrictions d’origine et de cookies pour les mutations, les paramètres liés SQL et les vérifications de scopes. Vérifier les réponses privées et leur cache au CDN, les permissions de suppression et l’absence de secrets dans les logs, snapshots partagés, variables VITE et bundles publics. Pour les déploiements AGPL, fournir aux utilisateurs le code source correspondant à la révision servie et conserver LICENSE et NOTICE.

## Budget mensuel indicatif

Un blog personnel non commercial utilisant 1 Go SQL et 3 Go-mois de médias, avec 1 million de lignes lues, 50 000 lignes écrites, 10 000 opérations R2 A et 100 000 opérations B, peut rester à 0 USD/mois tant que les autres quotas Hobby sont respectés. Compter les scans vectoriels, les logs, les sauvegardes et les retries. L’inférence dépend du fournisseur déjà choisi par l’opérateur. Turso Free indique 5 Go, 500 millions de lignes lues et 10 millions écrites par mois. R2 Standard inclut 10 Go-mois, un million d’opérations A et dix millions B. [Turso](https://turso.tech/pricing) · [R2](https://developers.cloudflare.com/r2/pricing/) · [Vercel](https://vercel.com/docs/functions/limitations)

## Jalons et critères de recette

- [x] Comparer la révision locale à l’amont et inventorier les dépendances au processus et au disque.
- [ ] Reproduire les builds Vue et Go/Linux CGO, exécuter les suites de référence et le conteneur local.
- [ ] Valider libSQL sur serveur jetable avec deux pools indépendants, migrations concurrentes, rollback, contraintes, index et distances L2.
- [ ] Brancher le handler Go sur le graphe commun, sans listener, migration concurrente ni travail dépendant de la survie du processus. Mesurer le bundle Linux, les démarrages à froid et les délais.
- [ ] Tester PUT direct et GET privé sur deux comptes, double finalisation, taille et type incorrects, avatar, Range vidéo et suppression.
- [ ] Persister état d’authentification, jobs, événements et statistiques. Interrompre un consommateur entre réservation et validation, puis vérifier la reprise.
- [ ] Restaurer un snapshot sur une cible vide, vérifier DB, médias et autorisations, puis comparer les API au serveur local et au catalogue OpenAPI.
- [ ] Assembler les règles Vercel et l’exemple d’environnement, passer just check et tester dans une preview après autorisation.

Pour synchroniser l’amont, intégrer les révisions sur une branche de travail, vérifier les nouveaux modèles, routes, effets asynchrones et accès disque, puis répéter la matrice SQLite/libSQL. Mettre à jour la révision de référence après la recette.
