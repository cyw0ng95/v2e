Feature: ProviderFSM Error Conditions

  Scenario: Start from non-IDLE state fails
    Given the provider "cve-provider" is in "RUNNING" state
    When StartProvider is called with provider ID "cve-provider"
    Then an error "invalid state transition" should be returned
    And the provider "cve-provider" state should remain "RUNNING"

  Scenario: Pause from non-RUNNING state fails
    Given the provider "cve-provider" is in "IDLE" state
    When PauseProvider is called with provider ID "cve-provider"
    Then an error "invalid state transition" should be returned
    And the provider "cve-provider" state should remain "IDLE"

  Scenario: Resume from non-PAUSED state fails
    Given the provider "cve-provider" is in "RUNNING" state
    When ResumeProvider is called with provider ID "cve-provider"
    Then an error "invalid state transition" should be returned
    And the provider "cve-provider" state should remain "RUNNING"

  Scenario: Stop from TERMINATED state fails
    Given the provider "cve-provider" is in "TERMINATED" state
    When StopProvider is called with provider ID "cve-provider"
    Then an error "invalid state transition" should be returned
    And the provider "cve-provider" state should remain "TERMINATED"
