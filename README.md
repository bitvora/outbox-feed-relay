# Outbox Feed Relay

Outbox Feed Relay generates a follow feed for every npub, using the outbox model to ensure notes are found more consistently.

It's built on the [Khatru](https://khatru.nostr.technology) framework.

## Prerequisites

- **Go**: Ensure you have Go installed on your system. You can download it from [here](https://golang.org/dl/).
- **Build Essentials**: If you're using Linux, you may need to install build essentials. You can do this by running `sudo apt install build-essential`.

## Setup Instructions

Follow these steps to get the outbox-feed-relay running on your local machine:

### 1. Clone the repository

```bash
git clone https://github.com/bitvora/outbox-feed-relay.git
cd outbox-feed-relay
```

### 2. Copy `.env.example` to `.env`

You'll need to create an `.env` file based on the example provided in the repository.

```bash
cp .env.example .env
```

### 3. Set your environment variables

Open the `.env` file and set the necessary environment variables. Example variables include:

```bash
RELAY_PUBKEY="e2ccf7cf20403f3f2a4a55b328f0de3be38558a7d5f33632fdaaefc726c1c8eb"
RELAY_ICON="https://pfp.nostr.build/d8fb3b6100a0eb9e652bbc34a0c043b7f225dc74e4ed6d733d0e059f9bd444d4.jpg"
```

### 4. Set Discovery Relays

Open the `discovery_relays.json` file and add pubkeys to the array

```json
[
  "wss://relay.damus.io",
  "wss://wot.utxo.one",
  "wss://nos.lol",
  "wss://relay.primal.net"
]
```

### 5. Build the project

Run the following command to build the relay:

```bash
go build
```

### 6. Create a Systemd Service (optional)

To have the relay run as a service, create a systemd unit file.

1. Create the file:

```bash
sudo nano /etc/systemd/system/outbox-feed-relay.service
```

2. Add the following contents:

```ini
[Unit]
Description=outbox-feed-relay Relay Service
After=network.target

[Service]
ExecStart=/home/ubuntu/outbox-feed-relay/outbox-feed-relay
WorkingDirectory=/home/ubuntu/outbox-feed-relay
Restart=always

[Install]
WantedBy=multi-user.target
```

3. Reload systemd to recognize the new service:

```bash
sudo systemctl daemon-reload
```

4. Start the service:

```bash
sudo systemctl start outbox-feed-relay
```

5. (Optional) Enable the service to start on boot:

```bash
sudo systemctl enable outbox-feed-relay
```

### 7. Serving over nginx (optional)

You can serve the relay over nginx by adding the following configuration to your nginx configuration file:

```nginx
server {
    listen 80;
    server_name yourdomain.com;

    location / {
        proxy_pass http://localhost:3334;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

Replace `yourdomain.com` with your actual domain name.

After adding the configuration, restart nginx:

```bash
sudo systemctl restart nginx
```

### 8. Install Certbot (optional)

If you want to serve the relay over HTTPS, you can use Certbot to generate an SSL certificate.

```bash
sudo apt-get update
sudo apt-get install certbot python3-certbot-nginx
```

After installing Certbot, run the following command to generate an SSL certificate:

```bash
sudo certbot --nginx
```

Follow the instructions to generate the certificate.

### 8. Access the relay

Once everything is set up, the relay will be running on `localhost:3334` or your domain name if you set up nginx.
