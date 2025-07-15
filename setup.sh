#!/bin/bash

echo "�� Setting up Flying Cup..."

# Check if .env file exists
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "✅ .env file created!"
    echo "⚠️  Please edit .env file with your actual configuration values"
else
    echo "✅ .env file already exists"
fi

# Create repos directory if it doesn't exist
if [ ! -d repos ]; then
    echo "📁 Creating repos directory..."
    mkdir -p repos
    echo "✅ repos directory created!"
else
    echo "✅ repos directory already exists"
fi

echo "🎉 Setup complete!"
echo ""
echo "Next steps:"
echo "1. Edit .env file with your configuration"
echo "2. Run: docker-compose up -d"
echo "3. Check logs: docker-compose logs -f"
