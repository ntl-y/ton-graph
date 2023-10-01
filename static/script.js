
document.getElementById("walletForm").addEventListener("submit", function(event) {
    event.preventDefault();
    var addressInput = document.getElementById("addressInput").value;

    fetch("/graph", {
        method: "POST",
        body: new URLSearchParams({ "addressInput": addressInput }),
        headers: {
            "Content-Type": "application/x-www-form-urlencoded"
        }
    })
    .then(response => response.json())
    .then(data => {
        // Очистить старый граф, удалив все элементы SVG
        d3.select("#resultOutput").selectAll("*").remove();
        // Создать массив уникальных узлов
        const nodesMap = {};
        // Specify the color scale.
        const color = d3.scaleOrdinal(d3.schemeCategory10);

        
        data.forEach(transaction => {
            nodesMap[transaction.From] = transaction.From;
            nodesMap[transaction.To] = transaction.To;
        });

        const nodes = Object.keys(nodesMap).map(node => ({ id: node }));
        // Создать связи (линии) на основе данных JSON
        const links = data.map(transaction => ({
            source: nodesMap[transaction.From],
            target: nodesMap[transaction.To]
        }));

        // Create a simulation with several forces.
        const simulation = d3.forceSimulation(nodes)
            .force("link", d3.forceLink(links).id(d => d.id))
            .force("charge", d3.forceManyBody())
            .force("x", d3.forceX())
            .force("y", d3.forceY());

        // Создать SVG контейнер
        const svg = d3.select("#resultOutput").append("svg")
            .attr("width", resultOutput.clientWidth)
            .attr("height", resultOutput.clientHeight)
            .attr("viewBox", [-resultOutput.clientWidth / 2, -resultOutput.clientHeight / 2, resultOutput.clientWidth, resultOutput.clientHeight]);

        // Add a line for each link, and a circle for each node.
        const link = svg.append("g")
            .attr("stroke", "#999")
            .attr("stroke-opacity", 0.6)
            .selectAll("line")
            .data(links)
            .join("line")
            .attr("stroke-width", 2); // Установите ширину линий по вашему выбору

            const node = svg.append("g")
            .attr("stroke", "#fff")
            .attr("stroke-width", 1.5)
            .selectAll("circle")
            .data(nodes)
            .join("circle")
            .attr("r", 5)
            .attr("fill", d => color(d.group))
            .on("mouseover", function(event, d) {
                // При наведении курсора на узел показать его идентификатор
                d3.select(this)
                    .append("title")
                    .text(d => d.id);
            })
            .on("mouseout", function(event, d) {
                // При убирании курсора убрать подсказку
                d3.select(this).select("title").remove();
            });
        node.append("title")
            .text(d => d.id);

        // Add a drag behavior.
        node.call(d3.drag()
                .on("start", dragstarted)
                .on("drag", dragged)
                .on("end", dragended));
        
        // Set the position attributes of links and nodes each time the simulation ticks.
        simulation.on("tick", () => {
            link
                .attr("x1", d => d.source.x)
                .attr("y1", d => d.source.y)
                .attr("x2", d => d.target.x)
                .attr("y2", d => d.target.y);

            node
                .attr("cx", d => d.x)
                .attr("cy", d => d.y);
        });

        // Reheat the simulation when drag starts, and fix the subject position.
        function dragstarted(event) {
            if (!event.active) simulation.alphaTarget(0.3).restart();
            event.subject.fx = event.subject.x;
            event.subject.fy = event.subject.y;
        }

        // Update the subject (dragged node) position during drag.
        function dragged(event) {
            event.subject.fx = event.x;
            event.subject.fy = event.y;
        }

        // Restore the target alpha so the simulation cools after dragging ends.
        // Unfix the subject position now that it’s no longer being dragged.
        function dragended(event) {
            if (!event.active) simulation.alphaTarget(0);
            event.subject.fx = null;
            event.subject.fy = null;
        }
    })
    .catch(error => {
        console.error("Произошла ошибка при получении данных:", error);
    });
});
