Feature: ProviderFSM State Transitions
  Background:
    Given the provider FSM service is available

  Scenario: Start provider from IDLE
    Given the provider "cve-provider" is in "IDLE" state
    When StartProvider is called with provider ID "cve-provider"
    Then the provider "cve-provider" should transition to "ACQUIRING" state
    And eventually the provider "cve-provider" should be in "RUNNING" state

  Scenario: Pause running provider
    Given the provider "cve-provider" is in "RUNNING" state
    When PauseProvider is called with provider ID "cve-provider"
    Then the provider "cve-provider" should be in "PAUSED" state

  Scenario: Resume paused provider
    Given the provider "cve-provider" is in "PAUSED" state
    When ResumeProvider is called with provider ID "cve-provider"
    Then the provider "cve-provider" should be in "RUNNING" state

  Scenario: Stop running provider
    Given the provider "cve-provider" is in "RUNNING" state
    When StopProvider is called with provider ID "cve-provider"
    Then the provider "cve-provider" should be in "TERMINATED" state
