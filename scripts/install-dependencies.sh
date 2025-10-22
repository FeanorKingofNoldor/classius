#!/bin/bash
# Classius Development Dependencies Installation Script
# Sets up complete development environment

set -e

echo "🚀 Installing Classius Development Dependencies"
echo "=============================================="

# Detect OS
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "📦 Installing Linux dependencies..."
    
    # Update package manager
    sudo apt update
    
    # Core build tools
    sudo apt install -y build-essential git curl wget
    
    # Qt development
    echo "🖥️  Installing Qt development tools..."
    sudo apt install -y qt6-base-dev qt6-declarative-dev qt6-tools-dev
    sudo apt install -y qml6-module-qtquick-controls
    
    # Cross-compilation for ARM
    echo "🔧 Installing ARM cross-compilation tools..."
    sudo apt install -y gcc-arm-linux-gnueabihf g++-arm-linux-gnueabihf
    
    # Audio development (for whistle detection)
    echo "🎵 Installing audio development libraries..."
    sudo apt install -y libasound2-dev portaudio19-dev libfftw3-dev
    
    # C++ development tools
    echo "⚙️  Installing C++ development tools..."
    sudo apt install -y cmake clang-format doxygen
    
elif [[ "$OSTYPE" == "darwin"* ]]; then
    echo "📦 Installing macOS dependencies..."
    
    # Check if Homebrew is installed
    if ! command -v brew &> /dev/null; then
        echo "Installing Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi
    
    # Install Qt
    brew install qt6
    
    # Cross-compilation tools (using zig for cross-compilation)
    brew install zig
    
    # Audio libraries
    brew install portaudio fftw
    
    # Development tools
    brew install cmake clang-format doxygen
    
else
    echo "❌ Unsupported operating system: $OSTYPE"
    echo "Please install dependencies manually"
    exit 1
fi

# Go installation
echo "🐹 Installing Go..."
if ! command -v go &> /dev/null; then
    GO_VERSION="1.21.3"
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        wget -q "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz"
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"
        rm "go${GO_VERSION}.linux-amd64.tar.gz"
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        export PATH=$PATH:/usr/local/go/bin
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        brew install go
    fi
else
    echo "✅ Go already installed: $(go version)"
fi

# Go development tools
echo "🔧 Installing Go development tools..."
go install github.com/cosmtrek/air@latest  # Hot reload
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Python installation and setup
echo "🐍 Setting up Python environment..."
if ! command -v python3 &> /dev/null; then
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        sudo apt install -y python3 python3-pip python3-venv
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        brew install python3
    fi
fi

# Python development tools
python3 -m pip install --upgrade pip
python3 -m pip install fastapi uvicorn pytest black isort flake8 mypy
python3 -m pip install openai anthropic  # AI APIs
python3 -m pip install numpy scipy scikit-learn  # ML/DSP

# Node.js for documentation and tooling
echo "📦 Installing Node.js..."
if ! command -v node &> /dev/null; then
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
        sudo apt-get install -y nodejs
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        brew install node
    fi
fi

# Docker for development services
echo "🐳 Installing Docker..."
if ! command -v docker &> /dev/null; then
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Install Docker
        curl -fsSL https://get.docker.com -o get-docker.sh
        sudo sh get-docker.sh
        sudo usermod -aG docker $USER
        rm get-docker.sh
        
        # Install Docker Compose
        sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        sudo chmod +x /usr/local/bin/docker-compose
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "Please install Docker Desktop for Mac from: https://docs.docker.com/desktop/mac/install/"
    fi
fi

# Database tools
echo "🗄️  Installing database tools..."
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    sudo apt install -y postgresql-client redis-tools
elif [[ "$OSTYPE" == "darwin"* ]]; then
    brew install postgresql redis
fi

# Create development directories
echo "📁 Creating project structure..."
mkdir -p {src/{device/{ui,core,audio,tests},server/{cmd/{server,migrate,seed},internal/{handlers,models,services,db},ai,tests},shared/{proto,types}},build,tools/{hooks,scripts},docker}

echo ""
echo "✅ Dependencies installation complete!"
echo ""
echo "📋 Next steps:"
echo "  1. Restart your shell or run: source ~/.bashrc"
echo "  2. Run: make setup"
echo "  3. Start development: make dev"
echo ""
echo "🔧 Installed tools:"
echo "  - Qt6 for device UI development"
echo "  - Go $(go version 2>/dev/null | cut -d' ' -f3) for backend services"
echo "  - Python $(python3 --version | cut -d' ' -f2) for AI services"
echo "  - Docker for development services"
echo "  - Cross-compilation tools for ARM devices"
echo ""

# Check if running in Docker group (Linux only)
if [[ "$OSTYPE" == "linux-gnu"* ]] && ! groups $USER | grep -q docker; then
    echo "⚠️  Note: You may need to log out and back in for Docker permissions to take effect"
fi