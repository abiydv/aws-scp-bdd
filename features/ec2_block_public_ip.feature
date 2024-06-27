Feature: Block public IPs for ec2 instances 

  Scenario: Block public IP on ec2
    Given I want to launch 1 ec2 instance
    And use subnet "public-subnet" in vpc "my-vpc"
    And set flag "AssociatePublicIpAddress" to "true"
    And add tags:
        | key  | value | add_to_resources |
        | dept | blue  | instance, network-interface, volume |
        | cost | 001   | instance, network-interface, volume |
        | proj | one   | instance, network-interface, volume |
    When I launch the ec2 instance
    Then the response is "UnauthorizedOperation"

  Scenario: Allow public IP on ec2 with exempt tag
    Given I want to launch 1 ec2 instance
    And use subnet "public-subnet" in vpc "my-vpc"
    And set flag "AssociatePublicIpAddress" to "true"
    And add tags:
        | key  | value | add_to_resources |
        | dept | blue  | instance, network-interface, volume |
        | cost | 001   | instance, network-interface, volume |
        | proj | one   | instance, network-interface, volume |
        | exempt | public-ip-control | instance, network-interface, volume |
    When I launch the ec2 instance
    Then the response is "OK"
