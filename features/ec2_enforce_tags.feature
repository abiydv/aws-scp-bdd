Feature: EC2 instances must have basic tags

  Scenario: Block ec2 with no tags
    Given I want to launch 1 ec2 instance 
    And use subnet "private-subnet-1" in vpc "my-vpc"
    And add tags:
        | key | value | add_to_resources |
    When I launch the ec2 instance
    Then the response is "UnauthorizedOperation"

  Scenario: Block ec2 with no cost tag
    Given I want to launch 1 ec2 instance 
    And use subnet "private-subnet-1" in vpc "my-vpc"
    And add tags:
        | key  | value | add_to_resources |
        | Name | first | instance, network-interface, volume |
        | dept | blue  | instance, network-interface, volume |
        | proj | one   | instance, network-interface, volume |
    When I launch the ec2 instance
    Then the response is "UnauthorizedOperation"

  Scenario: Block ec2 with invalid cost tag
    Given I want to launch 1 ec2 instance 
    And use subnet "private-subnet-1" in vpc "my-vpc"
    And add tags:
        | key  | value | add_to_resources |
        | Name | first | instance, network-interface, volume |
        | dept | blue  | instance, network-interface, volume |
        | cost | 100   | instance, network-interface, volume |
        | proj | one   | instance, network-interface, volume |
    When I launch the ec2 instance
    Then the response is "TagPolicyViolation"

  Scenario: Allow ec2 with basic tags
    Given I want to launch 1 ec2 instance
    And use subnet "private-subnet-1" in vpc "my-vpc"
    And add tags:
        | key  | value | add_to_resources |
        | Name | first | instance, network-interface, volume |
        | dept | blue  | instance, network-interface, volume |
        | cost | 001   | instance, network-interface, volume |
        | proj | one   | instance, network-interface, volume |
    When I launch the ec2 instance
    Then the response is "OK"
