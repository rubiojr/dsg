# DSG - DataHub Schema Generator

## Overview

DSG (DataHub Schema Generator) is a command-line tool that uses AI to generate and manage DataHub dataset schemas. It leverages OpenAI's language models to create dataset schemas based on user descriptions, then posts them directly to DataHub.

## Features

- AI-assisted dataset schema generation
- Support for Azure OpenAI and OpenAI APIs
- Local history management of generated schemas
- Direct integration with DataHub REST API
- View, manage, and deploy past schema generations

## Installation

### Prerequisites

- Go 1.24 or higher
- DataHub instance with API access
- OpenAI API key or Azure OpenAI access

### Building from source

```bash
git clone https://github.com/rubiojr/dsg
cd dsg
go build
```

Or:

```bash
go install github.com/rubiojr/dsg@latest
```

to install the latest version.

## Usage

### Environment Setup

You can set the following environment variables or pass them as command-line flags:

```bash
# DataHub configuration
export DATAHUB_GMS_URL=http://localhost:8080
export DATAHUB_GMS_TOKEN="your-datahub-token"

# OpenAI configuration
export OPENAI_API_KEY="your-openai-api-key"
export OPENAI_MODEL="gpt-4o"  # or another model

# For Azure OpenAI
export OPENAI_USE_AZURE=true
export OPENAI_API_BASE="https://your-azure-openai-endpoint"
export AZURE_OPENAI_DEPLOYMENT="deployment-name"
export AZURE_OPENAI_API_VERSION="2024-08-01-preview"
```

### Basic Commands

#### Generate a Dataset Schema

```bash
dsg generate
```

This will open an interactive prompt where you can describe the dataset you want to create. After writing your description, press Ctrl+D to submit. The AI will generate a schema and post it to DataHub automatically.

Generate using a previously used prompt:

```bash
dsg generate --prompt-from <ID> # see history command
```

#### View Generation History

```bash
dsg history
```

Shows a list of previously generated schemas.

#### View Details of a Specific Generation

```bash
dsg show 1  # Show details for history ID 1
```

#### Post an Existing Schema to DataHub

```bash
dsg post 1  # Post schema with history ID 1 to DataHub
```

#### Delete a History Entry

```bash
dsg delete 1  # Delete history entry with ID 1
```

#### Clear All History

```bash
dsg clear
```

## Examples

### Generating a Customer Dataset

```bash
$ dsg generate
Creating a DataHub dataset using NLP...

Writing temp prompt file to /tmp/XXXXXprompt...
Write the input for AI, hit Ctrl-D when finished:

Create a dataset named "customer_profiles" with fields for:
- customer_id (unique identifier)
- first_name
- last_name
- email
- signup_date
- last_purchase_date
- loyalty_tier (bronze, silver, gold, platinum)
- total_spend (numerical value)
- preferred_payment_method
^D

Understood!
Processing input and generating the dataset (may take a while)...
 🤖 finished!
1 datasets created! ☑
```

## License

[MIT](/LICENSE)
