const vm = require('vm');
const bodyParser = require('body-parser');
const express = require('express');
const app = express();

app.use(bodyParser.json());

app.post('/execute', (req, res) => {
	const code = req.body.code;
	try {
		const vmResult = vm.runInNewContext(code);
		res.send(vmResult);
	} catch (e) {
		res.send(e.toString());
	}
});

app.get('/health', (req, res) => {
	res.send(true);
});

const PORT = 8099;
app.listen(PORT, () => {
	console.log(`Server listening on ${PORT}`);
});