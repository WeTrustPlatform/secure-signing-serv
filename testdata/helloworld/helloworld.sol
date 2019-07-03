pragma solidity ^0.5.8;

contract helloWorld {
	string public message;

	event Updated(address indexed sender, string message);

	constructor() public {
		message = 'hello world!';
	}

	function setMessage(string memory m) public {
		message = m;
		emit Updated(msg.sender, m);
	}

	function renderMessage() public view returns (string memory) {
		return message;
	}
}
