# Portage Vercel d’Ech0

État du chantier au 20 novembre 2026. La référence est `a9445bb44e762ca9101cae51ea1d31b070114e9c`, commune au fork et au HEAD amont lors de l’audit. Ce document distingue la faisabilité du résultat des tests. La branche ne constitue pas encore une distribution Vercel de production.

## Application actuelle et cible

`cmd/ech0/main.go` initialise le processus avant Cobra. `serve` construit le graphe Wire, amorce la configuration métier, lance le gestionnaire de travaux, le planificateur et le serveur TCP. Gin distribue les routes Huma et les routes brutes aux mêmes services Go. Le binaire inclut `template/dist`, produit par Vite depuis `web` avec Node 26 et pnpm 12.16.1. Le module Go demande 1.27.4. Les applications `hub` et `site` ont leurs propres builds.

La cible conserve la même origine HTTPS et le routeur Gin. Les assets immuables peuvent être servis par le CDN. Les pages dont les métadonnées sont rendues par Go, la découverte, OAuth, MCP et les routes API doivent passer par Go avant toute règle SPA. Le service Go écoute sur PORT et délègue à un handler HTTP qui réutilise le graphe d’injection à chaud. Les pages et les assets embarqués passent actuellement par ce service. Les données partagées vont dans libSQL, les octets dans un bucket S3 privé, les fichiers de travail bornés dans le répertoire temporaire. Le démarrage doit échouer explicitement si le stockage distant est indisponible. [Runtime Go](https://vercel.com/docs/functions/runtimes/go)

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
| Gros POST multipart des clients existants | Bloqué | Vercel rejette au-delà de 4,5 MB avant Go. Le navigateur peut transférer directement vers R2. Décision validée pour le fork : adapter ces clients au protocole direct R2. Les anciens binaires qui n’implémentent que le POST conservent cette limite. [Limites des Functions](https://vercel.com/docs/functions/limitations) |
| Jobs, Busen, webhooks, indexation | Adaptation | Insérer l’événement avec la mutation métier. Réserver le travail par écriture conditionnelle, bail expirant et identifiant de tentative. Persister les étapes et livraisons. Un timeout après réception d’un webhook impose une clé stable pour la déduplication côté destinataire. |
| Nettoyage et statistiques | Adaptation | Nettoyage toutes les 72 h et flush visiteurs toutes les 60 min dans internal/task/scheduled. Écrire des compteurs durables à la visite et planifier le nettoyage par échéance. |
| Snapshots à cron configurable même sans trafic | Adaptation | Hobby accepte un cron quotidien avec précision horaire. Un réveil quotidien peut traiter les échéances quotidiennes. Vercel Queues sur Hobby est autorisé pour réveiller la fonction entre deux crons. Un rattrapage tardif ne constitue pas un snapshot historique. Le snapshot doit encore être rendu durable. [Queues](https://vercel.com/docs/queues/pricing) [Cron Hobby](https://vercel.com/docs/cron-jobs/usage-and-pricing) |
| Export ZIP, téléchargement, import Ech0/Memos, Capsule et destinations fs/S3 | Adaptation | Les exporters appellent VACUUM INTO sur un fichier local avant l’archivage. Produire une vue cohérente du primaire distant et un manifeste des objets. Importer dans une génération isolée, valider puis basculer atomiquement. Tester avec médias locaux et S3. |
| Dashboard, logs historiques, live WS et SSE | Adaptation | Le dashboard interroge un fichier et s’abonne aux logs du processus. Persister le flux filtré avec rétention. Tester séparément l’upgrade Gorilla sur le runtime Go et les reconnexions multi-instances. La disponibilité des WebSockets Vercel sur tous les plans ne valide pas cet adaptateur. [WebSockets](https://vercel.com/docs/functions/websockets) |
| CLI, TUI, Capsule CLI, Docker, sites satellites | Compatible | Conserver les entrées natives et le volume data. Le TUI interactif s’exécute sur le poste de l’opérateur. Comparer les exports CLI avec le format admin. |

## Données, sécurité et exploitation

Le DSN local active WAL, busy_timeout=5000, synchronous=NORMAL et txlock=immediate. La configuration ne force pas les clés étrangères dans ce DSN. AutoMigrate précède quatorze migrateurs applicatifs. Leurs marqueurs sont lus puis écrits séparément. WithStopOnError arrête la boucle sans transmettre l’échec à son appelant. La reconstruction des colonnes temporelles consulte sqlite_master et PRAGMA table_info, recrée index et triggers et tente de changer foreign_keys dans la transaction. Ne pas lancer cette séquence sur plusieurs démarrages à froid. Préparer les versions de schéma hors requête avec verrou durable, enregistrer le succès dans la même transaction que les changements et vérifier foreign_key_check.

Pour migrer une installation, travailler d’abord sur un snapshot cohérent de test, migrer sa copie avec le binaire de la même révision, inventorier les fichiers par clé et empreinte et vérifier les références privées. Exporter schéma et lignes ordinaires vers une base libSQL neuve. Extraire explicitement les vecteurs sous sqlite-vec (ne pas importer ses tables internes) puis les convertir et comparer les voisins. Réconcilier nombres de lignes, relations, index, empreintes et ACL. Répéter un aller-retour vers une installation SQLite vierge avant la bascule avec suspension des écritures. Garder le snapshot source et définir comment rejouer les écritures postérieures avant tout retour arrière. La conversion du fichier SQLite et la reconstruction de vec0 ont un test aller-retour. Le transfert des objets, la validation des anciennes URL et la bascule en production nécessitent une recette séparée.

Le driver officiellement indiqué pour le moteur libSQL distant est libsql-client-go. Le pilote tursogo-serverless cible un autre moteur. Séparer URL et jeton, exiger TLS en production et éviter d’inclure les identifiants dans les erreurs. [Référence Go Turso](https://docs.turso.tech/sdk/go/reference)

Préparer un bucket R2 Standard privé, une clé S3 limitée à ce bucket et un endpoint de compte. Autoriser uniquement l’origine HTTPS de l’application dans CORS, les méthodes et les en-têtes signés réellement nécessaires. Une URL PUT peut être réutilisée jusqu’à expiration. Valider et copier une version déterminée avant publication, purger les objets temporaires expirés. Pour les téléchargements, résoudre l’autorisation de lecture à chaque émission de GET signé. Conserver l’URL applicative stable pour les médias publics. [URLs présignées R2](https://developers.cloudflare.com/r2/api/s3/presigned-urls/)

Conserver le filtrage SSRF de pkg/egress pour Connect et les webhooks, l’échappement Markdown côté rendu, les restrictions d’origine et de cookies pour les mutations, les paramètres liés SQL et les vérifications de scopes. Vérifier les réponses privées et leur cache au CDN, les permissions de suppression et l’absence de secrets dans les logs, snapshots partagés, variables VITE et bundles publics. Pour les déploiements AGPL, fournir aux utilisateurs le code source correspondant à la révision servie et conserver LICENSE et NOTICE.

## Budget mensuel indicatif

Un blog personnel non commercial utilisant 1 Go SQL et 3 Go-mois de médias, avec 1 million de lignes lues, 50 000 lignes écrites, 10 000 opérations R2 A et 100 000 opérations B, peut rester à 0 USD/mois tant que les autres quotas Hobby sont respectés. Compter les scans vectoriels, les logs, les sauvegardes et les retries. L’inférence dépend du fournisseur déjà choisi par l’opérateur. Turso Free indique 5 Go, 500 millions de lignes lues et 10 millions écrites par mois. R2 Standard inclut 10 Go-mois, un million d’opérations A et dix millions B. [Turso](https://turso.tech/pricing) · [R2](https://developers.cloudflare.com/r2/pricing/) · [Vercel](https://vercel.com/docs/functions/limitations)


## État vérifié sur la branche

| Domaine | Résultat observé | Reste à vérifier ou à réaliser |
|---|---|---|
| Service Go | Deux processus Linux distincts servent le frontend Vue embarqué, le catalogue et OpenAPI. Le second voit le propriétaire créé par le premier et accepte sa connexion. Une route de découverte inconnue renvoie 404. | Build distant Vercel, temps à froid, WebSocket, longues réponses SSE, durée maximale de connexion. Deux demandes concurrentes avec des noms différents ont créé un seul propriétaire en recette locale. |
| Base | Deux pools contre un sqld jetable 0.24.32. Préparation concurrente du schéma, écritures simultanées, annulation d’une transaction, index, recherche LIKE et filtrage des vecteurs testés. Le schéma distant est versionné. Les migrations se lancent avant le trafic. | Tester contre Turso hébergé, avec les limites de durée et de volume du service, et mesurer le coût du scan vectoriel exact. |
| Identité | Révocations JWT, consommation unique des codes MCP, challenges WebAuthn, échange social, device flow, jetons de captcha et quotas des routes protégées par RateLimit partagent libSQL. Le test HTTP approuve le CLI sur une instance et récupère le jeton une seule fois sur l’autre. Les données de challenges sont chiffrées et les lectures mutables contournent le cache local. | Recette navigateur OIDC, passkeys HTTPS et captcha complet sur le domaine cible. |
| Médias | PUT direct signé pendant dix minutes, création conditionnelle, contrôle des octets, de la taille et du propriétaire avant publication. L’URL applicative /media/:key contrôle la visibilité et redirige vers une lecture courte du bucket privé. Le navigateur, les avatars, le logo et le CLI utilisent le flux direct en mode Vercel. | Recette sur R2 réel avec deux comptes, Range vidéo, révocation d’accès et suppression. Les instances chaudes relisent les réglages S3 avant les requêtes API et médias. La maintenance horaire reprend les suppressions des intentions expirées dans le bucket actif. Les fichiers temporaires déjà finalisés demandent encore une purge transactionnelle. |
| Visiteurs et Queues | Déduplication par empreinte HMAC, compteurs et historique partagés. Le cron réconcilie 26 créneaux horaires. Son consommateur attend la purge Go avant d’acquitter le message. | Livraisons Vercel réelles et cadence des snapshots. |
| Sauvegardes | Export libSQL dans une transaction, reconstruction SQLite native avec vec0. Le test relit un voisin à distance 5 puis reconvertit le fichier pour libSQL. | Le panneau produit un ZIP natif par blocs et parties S3 dans des invocations distinctes. Le test HTTP restaure un fichier de plus de 10 Mio après un échec de partie et une livraison en double. Le gros import web et la restauration complète des objets restent à réaliser. L’import métier Ech0 sélectionne des domaines et ne restaure pas à lui seul une instance complète. |
| Client et build | Build Vue de production, i18n, 176 tests Vitest et 135 tests CLI sous Linux. Un test CLI envoie une vidéo de 5 Mio vers un serveur objet distinct et vérifie que le jeton applicatif reste sur l’origine API. Build Go Linux CGO avec frontend embarqué d’environ 65 Mio. | Docker et environnements hébergés. Le poids du binaire local ne donne pas le poids du bundle final de Vercel. |

## Reproduire les contrôles

À la racine de la branche feat/vercel, utiliser Go 1.27.4, un compilateur C, les en-têtes SQLite, Node 26 et pnpm 12.16.1. Le répertoire template/dist doit exister avant les builds Go. Le script de build Vercel récupère les en-têtes correspondant à la version verrouillée de go-sqlite3.

```sh
pnpm -C web install --frozen-lockfile
pnpm -C web test:unit
pnpm -C web build
npm ci --prefix deployment/queue
node --test scripts/vercel/schedule.test.js
export ECH0_TEST_SQLD=/chemin/absolu/vers/sqld
go test ./...
go vet ./...
go build ./...
just check
VERCEL_OUTPUT_FILE="$PWD/server" sh scripts/vercel/build.sh
go build -o ./ech0 ./cmd/ech0
python3 scripts/vercel/smoke.py --sqld "$ECH0_TEST_SQLD" --cli "$PWD/ech0" --server "$PWD/server"
```

Le test sqld et le test HTTP lancent leurs propres bases dans des répertoires temporaires. Sans ECH0_TEST_SQLD, la suite Go omet le test d’intégration distant. Le test HTTP couvre le démarrage et la connexion sur deux processus. Son serveur S3 de test reçoit effectivement les octets de staging, simule un échec de partie, assemble le multipart et sert le ZIP privé. La recette sur Cloudflare reste ouverte.

## Installer une instance de recette

1. Partir de la branche feat/vercel. Créer une base **libSQL** de recette dans Turso, relever son URL et créer un jeton limité à cette base. Fournir les secrets uniquement au processus qui les utilise. Sur une base vide, exécuter `ech0 prepare-libsql`. Réexécuter cette commande avant de servir une nouvelle version du fork. Le cold start vérifie la version attendue. [Turso CLI](https://docs.turso.tech/cli/db/create)
2. Créer un bucket R2 Standard privé et une clé S3 en lecture/écriture restreinte au bucket. Laisser les accès publics désactivés. Autoriser uniquement les origines HTTPS de recette dans CORS, avec PUT, GET, HEAD et les en-têtes Content-Type, If-None-Match et Range. Le navigateur ajoute Content-Length au PUT. Depuis un navigateur de cette origine, contrôler le preflight et les en-têtes de la requête signée. [CORS R2](https://developers.cloudflare.com/r2/buckets/cors/) · [URL présignées](https://developers.cloudflare.com/r2/api/s3/presigned-urls/)
3. Utiliser la racine du dépôt comme racine de projet Vercel. vercel.json déclare le service Go backend, le service Node queue, la liaison interne ECH0_SERVICE_URL, le routage et le cron quotidien. Renseigner les valeurs de .env.example dans l’environnement Preview, avec les secrets distincts de la production. Le préfixe VITE rend une variable publique. Les identifiants du bucket et de la base sont exclusivement côté serveur. [Services Vercel](https://vercel.com/docs/services) · [Runtime Go](https://vercel.com/docs/functions/runtimes/go)
4. En preview autorisée, vérifier /healthz, /api/openapi.json, /, une route SPA et une route inconnue sous /.well-known. Créer le premier compte sur une instance de recette à accès restreint. Tester ensuite le même compte après un démarrage à froid, le téléchargement privé, l’envoi direct et la livraison effective de la queue. Les cases ouvertes ci-dessous conditionnent la mise en production.

| Variables | Valeur et rôle |
|---|---|
| ECH0_DB_TYPE, TURSO_DATABASE_URL, TURSO_AUTH_TOKEN | libsql, URL de la base et jeton de données. HTTP sans jeton est accepté uniquement sur les IP de loopback. |
| JWT_SECRET, CRON_SECRET | Deux secrets aléatoires distincts. Le premier doit rester stable entre instances et versions. Le second protège le cron et son worker. |
| ECH0_SETTING_SERVER_URL, ECH0_AUTH_WEBAUTHN_RP_ID, ECH0_AUTH_WEBAUTHN_ORIGINS | URL publique HTTPS, nom de domaine sans schéma, origines HTTPS autorisées séparées par des virgules. Enregistrer également les callbacks chez les fournisseurs d’identité. |
| ECH0_OBJECT_ENABLED, ECH0_UPLOAD_VERIFIED_DIRECT | true dans le service Go. Le build Vue Vercel active VITE_ECH0_OBJECT_UPLOADS. |
| ECH0_S3_PROVIDER, ECH0_S3_REGION, ECH0_S3_BUCKET, ECH0_S3_ENDPOINT | r2, auto, nom du bucket et endpoint S3 du compte Cloudflare. |
| ECH0_S3_ACCESS_KEY, ECH0_S3_SECRET_KEY, ECH0_S3_USE_SSL, ECH0_S3_USE_PATH_STYLE | Identifiants S3 dédiés au bucket, TLS et adressage par chemin. Les coordonnées persistées dans les réglages peuvent primer sur les valeurs initiales. |
| ECH0_LOG_FILE_ENABLE, ECH0_STORAGE_DATA_ROOT | false et /tmp/ech0-files. Le répertoire temporaire appartient à l’instance. |
| ECH0_SERVICE_URL | Fourni au consommateur par la liaison de services Vercel. |

## Migrer une base SQLite et répéter une restauration

1. Travailler sur une copie cohérente de l’installation de la révision auditée. Conserver la base, les médias et les paramètres de l’instance source. Le convertisseur vérifie le marqueur de migration amont, ouvre la source en lecture seule et refuse de remplacer un fichier existant.
2. Exécuter `ech0 convert-libsql /copie/ech0.db /copie/import-turso.db`. Le convertisseur inclut les écritures WAL validées, transforme vec0 en F32_BLOB et contrôle l’intégrité et les clés étrangères.
3. Pour un fichier dans la limite documentée de 2 Go, créer une nouvelle base de recette avec `turso db create ech0-recette --from-file /copie/import-turso.db`. Connecter le fork à ses propres identifiants et lancer `ech0 prepare-libsql`. L’import initial au-delà de cette limite demande encore un transfert par lots. [Import de fichier Turso](https://docs.turso.tech/cli/db/create)
4. Pour une installation avec médias locaux, inventorier clés, chemins résolus, empreintes et références des fichiers, avatars et logos. Copier les octets dans le bucket privé au chemin attendu par le schéma de stockage. Réconcilier les métadonnées de stockage et tester les anciennes URL avec un visiteur et avec leur propriétaire avant toute bascule. Le convertisseur ne déplace pas les objets. [Guide de stockage amont](usage/storage-migration.md)
5. Pour une vérification indépendante, `ech0 backup-libsql /copie/retour.db` produit un nouveau fichier SQLite natif. Démarrer le binaire local de recette sur une **copie** de retour.db, avec ECH0_DB_TYPE=sqlite et ECH0_DB_PATH pointant vers cette copie. Rétablir les médias à partir de leur propre sauvegarde privée et vérifier données, recherche et visibilité. La base contient les réglages et références des objets. Restreindre l’accès aux copies qui peuvent contenir des secrets.
6. Avant la bascule, suspendre les écritures de la source, répéter la copie finale et conserver la source intacte jusqu’à la vérification des totaux par table, relations, médias et autorisations. Après des écritures sur la cible, un retour arrière exige de les réconcilier.

## Sauvegarde web et traitement durable

Le panneau d’administration reste le point d’entrée pour toute taille prise en charge par les services. L’opérateur a choisi de conserver temporairement les anciennes versions des lignes **dans la base actuelle**. L’état et les baux se trouvent dans Turso. La suite des étapes est publiée dans Queues. Fermer le navigateur laisse le travail dans la base. Le cron réconcilie les travaux qui auraient manqué leur premier réveil. Les commandes du fork servent aussi aux vérifications indépendantes.

La capture réserve une génération unique dans le primaire et installe des triggers SQLite qui retiennent la première valeur modifiée ou supprimée dans chaque table. Les lignes ajoutées après le début sont exclues. L’expiration est prolongée après chaque lot réussi. Une modification du schéma pendant la capture invalide la génération. La préparation regroupe les DDL en batch. Turso documente une fenêtre de cinq secondes pour une transaction HTTP interactive. [SQL over HTTP](https://docs.turso.tech/sdk/http/reference)

Chaque worker lit jusqu’à 64 lignes, ouvre une vue privée du fichier SQLite natif et écrit les blocs modifiés de 64 Kio sous le préfixe privé ech0-staging/. Le commit du curseur et des références aux nouveaux blocs exige toujours le bail courant. Les blocs portent une empreinte SHA-256 qui est vérifiée à la lecture. La dernière étape reconstitue un ZIP contenant ech0.db avec CRC et extension ZIP64 pour les grands fichiers. Les parties S3 font au moins 8 Mio, sauf la dernière. Un upload terminé devient téléchargeable après publication de son nom en base. Le navigateur suit directement l’URL signée, avec une durée de cinq minutes. Les objets médias externes conservent la sémantique amont des snapshots S3 et doivent être sauvegardés séparément. [Multipart R2](https://developers.cloudflare.com/r2/objects/upload-objects/)

Les expressions de cron de cinq et six champs utilisent UTC comme le planificateur amont. Un réveil de réconciliation est prévu à chaque heure. Une occurrence proche reçoit son propre message différé. Une modification des réglages invalide les anciens messages dont l’expression diffère. Le traitement tolère les doublons. Prévoir dans le budget la taille de l’image intermédiaire, les anciennes versions de lignes, les opérations R2 de blocs, les messages par lot et jusqu’à trois archives complètes. Le gratuit dépend du taux de modification et de la fréquence réelle. Une livraison tardive capture l’état au moment de la capture. Le service ne reconstruit pas rétroactivement un état qui n’a jamais été capturé. [Queues et quotas](https://vercel.com/docs/queues/pricing)

Configurer une règle R2 d’abandon des uploads multipart incomplets pour couvrir la fenêtre entre l’allocation d’un upload et l’enregistrement de son identifiant. Le worker gère la purge des blocs de staging après succès, annulation ou échecs répétés. Une exception qui survient immédiatement après la création d’un objet peut laisser un bloc à examiner. Vérifier l’inventaire du préfixe et l’absence de capture active avant toute suppression manuelle. Le test ZIP64 valide les offsets au-delà de 4 Gio sur un lecteur clairsemé. Les limites des quotas réels et les performances à plusieurs gigaoctets demandent une preview.

## Jalons restant avant la production

- [x] Auditer les entrées, accès disque, workers, bus, dépendances SQLite et flux d’authentification.
- [x] Construire le frontend Vue et les exécutables Linux CGO, tester SQLite et libSQL localement, vérifier la connexion sur deux processus.
- [x] Adapter les clients web et CLI aux transferts directs et vérifier localement les signatures, la finalisation et l’isolation des jetons.
- [x] Vérifier une sauvegarde web multi-invocations, le rollback logique des écritures concurrentes, la relivraison, le retry multipart, la lecture du ZIP et le téléchargement privé.
- [x] Raccorder la rétention des statistiques au worker et tester la planification avec chevauchement.
- [x] Persister dans la même transaction les effets des écritures Echo et utilisateur. Le worker relit l’Echo courant pour l’indexer. Chaque webhook actif reçoit son propre enregistrement et le même identifiant à chaque essai. Un changement de destination annule la livraison en attente. Les tâches en cours sont exclues des snapshots.
- [x] Enregistrer les effets des commentaires, médias et snapshots. Les courriels de notification sont conservés dans une file privée distincte des webhooks. Le worker relit les paramètres SMTP et réutilise le même Message-ID lors des reprises. Un serveur SMTP peut délivrer deux fois après une réponse réseau ambiguë.
- [x] Reprendre et annuler une réindexation administrative depuis deux instances. Le test HTTP utilise un fournisseur d’embeddings local déterministe, vérifie le statut, la relivraison et la valeur réelle dans vec_echo.
- [ ] Porter les imports sur des étapes durables.
- [ ] Valider les volumes proches des quotas hébergés et achever les imports web de grosses archives Ech0, Memos et Capsule.
- [x] Acheminer les réponses aux confirmations Copilot entre instances via libSQL. Une écriture conditionnelle accepte une seule réponse du propriétaire. Les slots expirent avec le flux et sont exclus des snapshots.
- [x] Conserver les journaux en base avec la durée ECH0_LOG_FILE_MAX_AGE et filtrer les champs de secrets avant indexation. Les deux transports du panneau lisent le même curseur. Les tests locaux couvrent historique multi-processus, reprise SSE et réception WebSocket. Une instance interrompue avant le vidage du tampon peut perdre ses dernières lignes.
- [ ] Valider en navigateur les flux de sécurité et les reconnexions réelles.
- [ ] Tester Docker, la preview Hobby, R2 et Turso hébergés, les limites de durée, les streams, les budgets réels et l’ensemble des accès privés.

Pour synchroniser l’amont, intégrer les révisions sur une branche de travail, vérifier les nouveaux modèles, routes, effets asynchrones et accès disque, régénérer Wire et OpenAPI, puis répéter la matrice SQLite/libSQL. Mettre à jour la révision de référence après la recette.
