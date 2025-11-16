// Latest Anthropic Claude 4 models
client<llm> CustomOpus4 {
  provider anthropic
  options {
    model "claude-opus-4-1-20250805"
    api_key env.ANTHROPIC_API_KEY
  }
}

client<llm> CustomSonnet4 {
  provider anthropic
  options {
    model "claude-sonnet-4-20250514"
    api_key env.ANTHROPIC_API_KEY
  }
}

client<llm> CustomHaiku {
  provider anthropic
  retry_policy Constant
  options {
    model "claude-3-5-haiku-20241022"
    api_key env.ANTHROPIC_API_KEY
  }
}

client<llm> CustomOllamaGPT {
  provider openai-generic
  options {
    base_url "http://localhost:11434/v1"
    model "gpt-oss:120b-cloude"
    default_role "user"
  }
}

// https://docs.boundaryml.com/docs/snippets/clients/round-robin
client<llm> CustomFast {
  provider round-robin
  options {
    // This will alternate between the two clients
    strategy [CustomOllamaGPT, CustomHaiku]
  }
}

// https://docs.boundaryml.com/docs/snippets/clients/fallback
client<llm> AnthropicFallback {
  provider fallback
  options {
    // This will try the clients in order until one succeeds
    strategy [CustomHaiku, CustomSonnet4, CustomOpus4]
  }
}

// https://docs.boundaryml.com/docs/snippets/clients/retry
retry_policy Constant {
  max_retries 3
  strategy {
    type constant_delay
    delay_ms 200
  }
}

retry_policy Exponential {
  max_retries 2
  strategy {
    type exponential_backoff
    delay_ms 300
    multiplier 1.5
    max_delay_ms 10000
  }
}
